package main

import (
	"net/http"
	"os"
	"strings"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

type Fid2Cid struct {
	Fid int
	Cid string
}

const (
	TestResultDir    = "test_result"
	ImagesDir        = "images"
	CompareImagesDir = "compare_images"
	URandom          = "/dev/urandom"
	FakeFid          = -1
)

var (
	params Params
	input  Input

	chFid2Cids = make(chan Fid2Cid, 10000)
	chResults  = make(chan Result, 20000)

	logger *zap.Logger
)

func main() {
	logger, _ = zap.NewDevelopment()

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
			&cli.StringFlag{
				Name:        "test_result_file",
				Usage:       "the file To save test result",
				Destination: &params.TestResultFile,
				Aliases:     []string{"trf"},
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
			&cli.BoolFlag{
				Name:        "timestamp", // TODO remove and save timestamp to test result file
				Usage:       "add timestamp to test result file",
				Value:       true,
				Destination: &params.Timestamp,
				Aliases:     []string{"ts"},
			},
		},
		Before: func(context *cli.Context) error {
			e := input.check()
			if e != nil {
				return e
			}

			e = os.MkdirAll(TestResultDir, os.ModePerm)
			if e != nil {
				logger.Error("create Dir err", zap.String("err", e.Error()))
				return e
			}

			e = os.MkdirAll(ImagesDir, os.ModePerm)
			if e != nil {
				logger.Error("create Dir err", zap.String("err", e.Error()))
				return e
			}

			e = os.MkdirAll(CompareImagesDir, os.ModePerm)
			if e != nil {
				logger.Error("create Dir err", zap.String("err", e.Error()))
				return e
			}

			return nil
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
					return Compare(
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
						Name:   "pin",
						Before: loadTestCids,
						Subcommands: []*cli.Command{
							{
								Name: "get",
								Action: func(context *cli.Context) error {
									input.TestCase = "ClusterPinGet"
									return doRequests(http.MethodGet, "/pins/")
								},
							},
							{
								Name: "add",
								Action: func(context *cli.Context) error {
									input.TestCase = "ClusterPinAdd"
									return doRequests(http.MethodPost, "/pins/ipfs/")
								},
							},
							{
								Name: "rm",
								Action: func(context *cli.Context) error {
									input.TestCase = "ClusterPinRm"
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
							input.TestCase = "ClusterAdd"
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
						Action: func(context *cli.Context) error {
							input.TestCase = "ClusterUnpin"
							bs, e := File2bytes(context.String("cids_file"))
							if e != nil {
								return e
							}

							cids := strings.Split(strings.TrimSpace(string(bs)), "\n")
							go func() {
								for _, cid := range cids {
									if len(cid) > 10 {
										chFid2Cids <- Fid2Cid{Fid: FakeFid, Cid: cid}
									}
								}
								close(chFid2Cids)
							}()

							return doRequests(http.MethodDelete, "/pins/ipfs/")
						},
					},
				},
			},
			{
				Name: "ipfs",
				Subcommands: []*cli.Command{
					{
						Name:   "stat",
						Before: loadTestCids,
						Action: func(context *cli.Context) error {
							input.TestCase = "IpfsStat"
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
					input.TestCase = "GenFile"
					return genFiles()
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		logger.Fatal(err.Error())
	}
}
