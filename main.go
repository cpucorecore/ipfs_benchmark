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
	chResults  = make(chan Result, 30000)

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
				Value:       1000,
				Destination: &input.To,
			},
			&cli.BoolFlag{
				Name:        "verbose",
				Usage:       "Verbose log",
				Value:       false,
				Destination: &params.Verbose,
				Aliases:     []string{"v"},
			},
			&cli.BoolFlag{
				Name:        "sync",
				Usage:       "sync concurrent request number",
				Value:       false,
				Destination: &params.Sync,
				Aliases:     []string{"s"},
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
						Name:    "timestamp",
						Usage:   "add timestamp to the file name",
						Value:   true,
						Aliases: []string{"t"},
					},
					&cli.BoolFlag{
						Name:    "sort_tps",
						Usage:   "sort the window tps values",
						Value:   true,
						Aliases: []string{"st"},
					},
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
						context.Bool("sort_tps"),
						context.Bool("sort_latency"),
						context.Bool("timestamp"),
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
							return loadFid2CidsFromTestFile()
						},
						Subcommands: []*cli.Command{
							{
								Name: "get",
								Action: func(context *cli.Context) error {
									input.TestCase = TestCaseClusterPinGet
									return doHttpRequests(http.MethodGet, "/pins", true)
								},
							},
							{
								Name: "add",
								Flags: []cli.Flag{
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
									input.TestCase = TestCaseClusterPinAdd
									return doHttpRequests(http.MethodPost, "/pins/ipfs", true)
								},
							},
							{
								Name: "rm",
								Action: func(context *cli.Context) error {
									input.TestCase = TestCaseClusterPinRm
									return doHttpRequests(http.MethodDelete, "/pins/ipfs", true)
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
							return postFiles()
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
								Required:    true,
							},
						},
						Before: func(context *cli.Context) error {
							return loadFileCids(context.String("cids_file"))
						},
						Action: func(context *cli.Context) error {
							input.TestCase = TestCaseClusterPinRm
							return doHttpRequests(http.MethodDelete, "/pins/ipfs", true)
						},
					},
				},
			},
			{
				Name: "ipfs",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "test_case",
						Usage:       "test case name",
						Destination: &input.TestCase,
						Aliases:     []string{"tc"},
						Required:    true,
					},
					&cli.StringFlag{
						Name:        "api_path",
						Usage:       "api path, eg: [/api/v0/swarm/peers, /api/v0/id]",
						Destination: &input.ApiPath,
						Aliases:     []string{"ap"},
						Required:    true,
					},
					&cli.StringFlag{
						Name:        "http_method",
						Usage:       "http method: [GET/POST/DELETE]",
						Destination: &input.HttpMethod,
						Value:       "POST",
						Aliases:     []string{"hm"},
						Required:    true,
					},
				},
				Subcommands: []*cli.Command{
					{
						Name: "repeat_request",
						Flags: []cli.Flag{
							&cli.UintFlag{
								Name:        "repeat",
								Usage:       "repeat per goroutine",
								Destination: &input.Repeat,
								Value:       100,
								Aliases:     []string{"r"},
								Required:    true,
							},
						},
						Action: func(context *cli.Context) error {
							return doRequestsRepeat(input.HttpMethod, input.ApiPath, int(input.Repeat))
						},
					},
					{
						Name: "iter_request",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:        "test_result_file",
								Usage:       "the file To save test result",
								Destination: &params.TestResultFile,
								Aliases:     []string{"trf"},
								Required:    true,
							},
							&cli.BoolFlag{
								Name:        "drop",
								Usage:       "drop http response, for /api/v0/cat api",
								Destination: &params.Drop,
								Value:       true,
								Aliases:     []string{"d"},
							},
						},
						Before: func(context *cli.Context) error {
							return loadFid2CidsFromTestFile()
						},
						Action: func(context *cli.Context) error {
							return doHttpRequests(
								input.HttpMethod,
								input.ApiPath,
								false,
							)
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
