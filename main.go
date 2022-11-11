package main

import (
	"net/http"
	"os"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

const (
	DeviceURandom = "/dev/urandom"
)

const (
	TestCaseGenFile = "GenFile"

	TestCaseIpfsStat = "IpfsStat"

	TestCaseClusterAdd    = "ClusterAdd"
	TestCaseClusterPinGet = "ClusterPinGet"
	TestCaseClusterPinAdd = "ClusterPinAdd"
	TestCaseClusterPinRm  = "ClusterPinRm"
)

var (
	params Params
	input  Input

	chFid2Cids = make(chan Fid2Cid, 20000)
	chResults  = make(chan Result, 20000)

	logger *zap.Logger
)

func init() {
	logger, _ = zap.NewDevelopment()

	ec := initDirs()
	if ec > 0 {
		logger.Error("initDirs failed", zap.Int("failed", ec))
		os.Exit(-1)
	}
}

func main() {
	app := &cli.App{
		Name: "ipfs_benchmark",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:        "goroutines",
				Usage:       "goroutine number",
				Value:       1,
				Destination: &input.Goroutines,
				Aliases:     []string{"g"},
			},
			&cli.IntFlag{
				Name:        "from",
				Usage:       "[from, to)",
				Value:       0,
				Destination: &input.From,
			},
			&cli.IntFlag{
				Name:        "to",
				Usage:       "[from, to)",
				Value:       100,
				Destination: &input.To,
			},
			&cli.IntFlag{
				Name:        "window",
				Usage:       "analyseResults Window",
				Value:       100,
				Destination: &params.Window,
				Aliases:     []string{"w"},
			},
			&cli.BoolFlag{
				Name:        "verbose",
				Usage:       "Verbose log",
				Value:       false,
				Destination: &params.Verbose,
				Aliases:     []string{"v"},
			},
			&cli.StringFlag{
				Name:        "host_port",
				Usage:       "api host and port",
				Value:       "192.168.0.85:9094",
				Destination: &input.HostPort,
				Aliases:     []string{"hp"},
			},
			&cli.StringFlag{
				Name:        "tag",
				Usage:       "add tag for test case, eg crdt, raft",
				Value:       "",
				Destination: &input.Tag,
				Aliases:     []string{"t"},
			},
		},
		Before: func(context *cli.Context) error {
			return input.check()
		},
		Commands: []*cli.Command{
			{
				Name:  "compare",
				Usage: "compare test result",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "sort_latency",
						Usage:   "sort the latencies",
						Value:   true,
						Aliases: []string{"sl"},
					},
				},
				Action: func(context *cli.Context) error {
					return CompareTests(
						context.String("tag"),
						context.Bool("sort_latency"),
						context.Int("window"),
						context.Args().Slice()...)
				},
			},
			{
				Name:  "gc",
				Usage: "cluster gc",
				Action: func(context *cli.Context) error {
					return gc()
				},
			},
			{
				Name: "cluster",
				Subcommands: []*cli.Command{
					{
						Name: "pin",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:        "test_result_file",
								Usage:       "the file To save test result",
								Destination: &params.TestResultFile,
								Aliases:     []string{"trf"},
								Required:    true,
							},
						},
						Before: func(context *cli.Context) error {
							return loadTestCids()
						},
						Subcommands: []*cli.Command{
							{
								Name: "get",
								Action: func(context *cli.Context) error {
									input.TestCase = TestCaseClusterPinGet
									return doRequests(http.MethodGet, "/pins/")
								},
							},
							{
								Name: "add",
								Action: func(context *cli.Context) error {
									input.TestCase = TestCaseClusterPinAdd
									return doRequests(http.MethodPost, "/pins/ipfs/")
								},
							},
							{
								Name: "rm",
								Action: func(context *cli.Context) error {
									input.TestCase = TestCaseClusterPinRm
									return doRequests(http.MethodDelete, "/pins/ipfs/")
								},
							},
						},
					},
					{
						Name: "add",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:        "files_dir",
								Value:       "files",
								Destination: &params.FilesDir,
								Aliases:     []string{"d"},
							},
							&cli.IntFlag{
								Name:        "block_size",
								Usage:       "block size, max value 1048576(1MB), default 262144(256KB)",
								Destination: &input.BlockSize,
								Value:       1024 * 256,
								Aliases:     []string{"bs"},
							},
							&cli.BoolFlag{
								Name:        "pin",
								Value:       true,
								Destination: &input.Pin,
								Aliases:     []string{"p"},
							},
							&cli.IntFlag{
								Name:        "replication_min",
								Value:       2,
								Destination: &input.ReplicationMin,
								Aliases:     []string{"rmin"},
							},
							&cli.IntFlag{
								Name:        "replication_max",
								Value:       2,
								Destination: &input.ReplicationMax,
								Aliases:     []string{"rmax"},
							},
						},
						Action: func(context *cli.Context) error {
							input.TestCase = TestCaseClusterAdd
							return sendFiles()
						},
					},
					{
						Name:  "unpin_by_cids",
						Usage: "unpin by cids list file",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:        "cids_file",
								Value:       "cids list file",
								Destination: &params.FilesDir,
								Aliases:     []string{"c"},
							},
						},
						Before: func(context *cli.Context) error {
							return loadFileCids(context.String("cids_file"))
						},
						Action: func(context *cli.Context) error {
							input.TestCase = TestCaseClusterPinRm
							return doRequests(http.MethodDelete, "/pins/ipfs/")
						},
					},
				},
			},
			{
				Name: "ipfs",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "test_result_file",
						Usage:       "the file To save test result",
						Destination: &params.TestResultFile,
						Aliases:     []string{"trf"},
						Required:    true,
					},
				},
				Subcommands: []*cli.Command{
					{
						Name: "stat",
						Before: func(context *cli.Context) error {
							return loadTestCids()
						},
						Action: func(context *cli.Context) error {
							input.TestCase = TestCaseIpfsStat
							return doRequests(http.MethodPost, "/api/v0/repo/stat/")
						},
					},
				},
			},
			{
				Name: "gen_files",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "files_dir",
						Value:       "files",
						Destination: &params.FilesDir,
						Aliases:     []string{"d"},
					},
					&cli.IntFlag{
						Name:        "size",
						Usage:       "file size",
						Value:       1024 * 1024,
						Destination: &params.FileSize,
						Aliases:     []string{"s"},
					},
				},
				Action: func(context *cli.Context) error {
					input.TestCase = TestCaseGenFile
					return genFiles()
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		logger.Fatal(err.Error())
	}
}
