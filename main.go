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

var (
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

var verbose bool

var goroutines int
var syncConcurrency bool

var dropHttpRespBody = false

var hostPort string

func main() {
	app := &cli.App{
		Name: "ipfs_benchmark",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "verbose",
				Usage:       "Verbose log",
				Value:       false,
				Destination: &verbose,
				Aliases:     []string{"v"},
			},
			&cli.IntFlag{
				Name:        "goroutines",
				Usage:       "goroutine number",
				Value:       1,
				Destination: &goroutines,
				Aliases:     []string{"g"},
			},
			&cli.BoolFlag{
				Name:        "sync_concurrency",
				Value:       false,
				Destination: &syncConcurrency,
				Aliases:     []string{"sc"},
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "compare",
				Usage: "compare test file",
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
					input := CompareInput{
						Timestamp:   context.Bool("timestamp"),
						Tag:         context.String("tag"),
						SortTps:     context.Bool("sort_tps"),
						SortLatency: context.Bool("sort_latency"),
					}
					return CompareTests(input, context.Args().Slice()...)
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
								Name:     "test_result_file",
								Usage:    "the file To save test result",
								Aliases:  []string{"trf"},
								Required: true,
							},
						},
						Before: func(context *cli.Context) error {
							return loadFid2CidsFromTestFile(context.String("test_result_file"))
						},
						Subcommands: []*cli.Command{
							{
								Name: "add",
								Flags: []cli.Flag{
									&cli.IntFlag{
										Name:  "replica",
										Value: 2,
									},
								},
								Action: func(context *cli.Context) error {
									input := ClusterPinAddInput{}
									input.Method = http.MethodPost
									input.Path = "/pins/ipfs"
									return doIterUrlHttpInput(input)
								},
							},
							{
								Name: "rm",
								Action: func(context *cli.Context) error {
									input := ClusterPinInput{}
									input.Method = http.MethodDelete
									input.Path = "/pins/ipfs"
									return doIterUrlHttpInput(input)
								},
							},
							{
								Name: "get",
								Action: func(context *cli.Context) error {
									input := ClusterPinInput{}
									input.Method = http.MethodGet
									input.Path = "/pins"
									return doIterUrlHttpInput(input)
								},
							},
						},
					},
					{
						Name: "add",
						Flags: []cli.Flag{
							&cli.IntFlag{
								Name:    "from",
								Value:   0,
								Aliases: []string{"f"},
							},
							&cli.IntFlag{
								Name:    "to",
								Value:   10,
								Aliases: []string{"t"},
							},
							&cli.IntFlag{
								Name:    "block_size",
								Usage:   "block size, max value 1048576(1MB), default 262144(256KB)",
								Value:   1024 * 256,
								Aliases: []string{"bs"},
							},
							&cli.BoolFlag{
								Name:    "pin",
								Value:   true,
								Aliases: []string{"p"},
							},
							&cli.IntFlag{
								Name:  "replica",
								Value: 2,
							},
						},
						Action: func(context *cli.Context) error {
							input := ClusterAddInput{}
							input.From = context.Int("from")
							input.To = context.Int("to")
							input.BlockSize = context.Int("block_size")
							input.Pin = context.Bool("pin")
							input.Replica = context.Int("replica")
							return postFiles(input)
						},
					},
					{
						Name:  "unpin_by_cids",
						Usage: "unpin by cids list file",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "cids_file",
								Value:    "cids list file",
								Aliases:  []string{"c"},
								Required: true,
							},
						},
						Before: func(context *cli.Context) error {
							return loadFileCids(context.String("cids_file"))
						},
						Action: func(context *cli.Context) error {
							input := ClusterPinInput{}
							input.Method = http.MethodDelete
							input.Path = "/pins/ipfs"
							return doIterUrlHttpInput(input)
						},
					},
				},
			},
			{
				Name: "ipfs",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "method",
						Usage:    "http method: [GET/POST/DELETE]",
						Aliases:  []string{"m"},
						Required: true,
					},
					&cli.StringFlag{
						Name:     "path",
						Usage:    "api path, eg: [/api/v0/swarm/peers, /api/v0/id]",
						Aliases:  []string{"p"},
						Required: true,
					},
				},
				Subcommands: []*cli.Command{
					{
						Name: "repeat_test",
						Flags: []cli.Flag{
							&cli.UintFlag{
								Name:     "repeat",
								Usage:    "repeat per goroutine",
								Aliases:  []string{"r"},
								Required: true,
							},
						},
						Action: func(context *cli.Context) error {
							input := IpfsRepeatInput{}
							input.Method = context.String("method")
							input.Path = context.String("path")
							input.Repeat = context.Int("repeat")
							return doRepeatHttpInput(input)
						},
					},
					{
						Name: "iter_test",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "test_result_file",
								Aliases:  []string{"trf"},
								Required: true,
							},
						},
						Before: func(context *cli.Context) error {
							return loadFid2CidsFromTestFile(context.String("test_result_file"))
						},
						Action: func(context *cli.Context) error {
							input := IpfsIterInput{}
							input.Method = context.String("method")
							input.Path = context.String("path")
							input.TestFile = context.String("test_result_file")
							return doIpfsIterInput(input)
						},
					},
				},
			},
			{
				Name: "gen_files",
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:    "from",
						Value:   0,
						Aliases: []string{"f"},
					},
					&cli.IntFlag{
						Name:    "to",
						Value:   10,
						Aliases: []string{"t"},
					},
					&cli.IntFlag{
						Name:    "size",
						Value:   1024 * 1024,
						Aliases: []string{"s"},
					},
				},
				Action: func(context *cli.Context) error {
					input := GenFileInput{}
					input.From = context.Int("from")
					input.To = context.Int("to")
					input.Size = context.Int("size")
					return genFiles(input)
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		logger.Fatal(err.Error())
	}
}
