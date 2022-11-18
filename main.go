package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

var (
	iInput IInput

	detail          bool
	goroutines      int
	syncConcurrency bool

	host string
	port string

	chFid2Cids = make(chan Fid2Cid, 20000)
	chResults  = make(chan Result, 30000)

	logger *zap.Logger
)

func baseParamsStr() string {
	return fmt.Sprintf("detail-%v_g-%d_sync-%v", detail, goroutines, syncConcurrency)
}

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
			&cli.BoolFlag{
				Name:        "detail",
				Usage:       "verbose log",
				Value:       false,
				Destination: &detail,
				Aliases:     []string{"d"},
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
			&cli.StringFlag{ // TODO not global params
				Name:        "host",
				Value:       "127.0.0.1",
				Destination: &host,
			},
			&cli.StringFlag{ // TODO not global params
				Name:        "port",
				Value:       "9094",
				Destination: &port,
				Aliases:     []string{"p"},
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "compare",
				Usage: "compare test file",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "tag", // TODO remove this flag, instead by infos from [test files...]
						Usage:    "compare tag",
						Aliases:  []string{"t"},
						Required: true,
					},
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
					var input CompareInput

					input.Tag = context.String("tag")
					input.Timestamp = context.Bool("timestamp")
					input.SortTps = context.Bool("sort_tps")
					input.SortLatency = context.Bool("sort_latency")

					if e := input.check(); e != nil {
						return e
					}

					iInput = input
					return CompareTests(input, context.Args().Slice()...)
				},
			},
			{
				Name: "cluster",
				Subcommands: []*cli.Command{
					{
						Name: "gc",
						Action: func(context *cli.Context) error {
							return gc(host + ":" + port)
						},
					},
					{
						Name: "pin",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "test_result_file",
								Usage:    "the file To save test result",
								Aliases:  []string{"trf"},
								Required: true,
							},
							&cli.IntFlag{
								Name:    "from",
								Aliases: []string{"f"},
								Value:   0,
							},
							&cli.IntFlag{
								Name:    "to",
								Aliases: []string{"t"},
								Value:   0,
							},
						},
						Before: func(context *cli.Context) error {
							return loadFid2CidsFromTestFile(
								context.String("test_result_file"),
								context.Int("from"),
								context.Int("to"),
							)
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
									var input ClusterPinAddInput

									input.Host = host
									input.Port = port
									input.Method = http.MethodPost
									input.Path = "/pins/ipfs"
									input.DropHttpResp = false
									input.TestFile = context.String("test_result_file")
									input.Replica = context.Int("replica")

									if e := input.check(); e != nil {
										return e
									}

									iInput = input
									return doIterHttpRequest(input)
								},
							},
							{
								Name: "rm",
								Action: func(context *cli.Context) error {
									var input ClusterPinRmInput

									input.Host = host
									input.Port = port
									input.Method = http.MethodDelete
									input.Path = "/pins/ipfs"
									input.DropHttpResp = false
									input.TestFile = context.String("test_result_file")

									if e := input.check(); e != nil {
										return e
									}

									iInput = input
									return doIterHttpRequest(input)
								},
							},
							{
								Name: "get",
								Action: func(context *cli.Context) error {
									var input ClusterPinGetInput

									input.Host = host
									input.Port = port
									input.Method = http.MethodGet
									input.Path = "/pins"
									input.DropHttpResp = false
									input.TestFile = context.String("test_result_file")

									if e := input.check(); e != nil {
										return e
									}
									iInput = input
									return doIterHttpRequest(input)
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
								Aliases: []string{"s"},
							},
							&cli.IntFlag{
								Name:    "replica",
								Value:   2,
								Aliases: []string{"r"},
							},
							&cli.BoolFlag{
								Name:    "pin",
								Value:   true,
								Aliases: []string{"p"},
							},
						},
						Action: func(context *cli.Context) error {
							var input ClusterAddInput

							input.Host = host
							input.Port = port
							input.Method = http.MethodPost
							input.Path = "/add"
							input.DropHttpResp = false
							input.From = context.Int("from")
							input.To = context.Int("to")
							input.BlockSize = context.Int("block_size")
							input.Replica = context.Int("replica")
							input.Pin = context.Bool("pin")

							if e := input.check(); e != nil {
								return e
							}

							iInput = input
							return postFiles(input)
						},
					},
					{
						Name:  "unpin_by_cid",
						Usage: "unpin by cids file",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "cid_file",
								Value:    "cids list file",
								Aliases:  []string{"c"},
								Required: true,
							},
						},
						Before: func(context *cli.Context) error {
							return loadFileCids(context.String("cid_file"))
						},
						Action: func(context *cli.Context) error {
							var input ClusterPinRmInput

							input.Host = host
							input.Port = port
							input.Method = http.MethodDelete
							input.Path = "/pins/ipfs"
							input.DropHttpResp = false
							input.TestFile = context.String("test_result_file")

							if e := input.check(); e != nil {
								return e
							}

							iInput = input
							return doIterHttpRequest(input)
						},
					},
				},
			},
			{
				Name: "ipfs",
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
						Subcommands: []*cli.Command{
							{
								// curl -X POST "http://127.0.0.1:5001/api/v0/swarm/peers?verbose=<value>&streams=<value>&latency=<value>&direction=<value>"
								Name: "swarm_peers",
								Flags: []cli.Flag{
									&cli.BoolFlag{
										Name:    "verbose",
										Aliases: []string{"v"},
										Value:   true,
									},
									&cli.BoolFlag{
										Name:  "streams",
										Value: true,
									},
									&cli.BoolFlag{
										Name:  "latency",
										Value: true,
									},
									&cli.BoolFlag{
										Name:  "direction",
										Value: true,
									},
								},
								Action: func(context *cli.Context) error {
									var input IpfsSwarmPeersInput

									input.Host = host
									input.Port = port
									input.Method = http.MethodPost
									input.Path = "/api/v0/swarm/peers"
									input.DropHttpResp = false
									input.Repeat = context.Int("repeat")
									input.Verbose = context.Bool("verbose")
									input.Streams = context.Bool("streams")
									input.Latency = context.Bool("latency")
									input.Direction = context.Bool("direction")

									if e := input.check(); e != nil {
										return e
									}

									iInput = input
									return doRepeatHttpInput(input)
								},
							},
							{
								// curl -X POST "http://127.0.0.1:5001/api/v0/id?arg=<peerid>&format=<value>&peerid-base=b58mh"
								Name: "id",
								Action: func(context *cli.Context) error {
									var input IpfsIdInput

									input.Host = host
									input.Port = port
									input.Method = http.MethodPost
									input.Path = "/api/v0/id"
									input.DropHttpResp = false
									input.Repeat = context.Int("repeat")

									if e := input.check(); e != nil {
										return e
									}

									iInput = input
									return doRepeatHttpInput(input)
								},
							},
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
							&cli.IntFlag{
								Name:    "from",
								Aliases: []string{"f"},
								Value:   0,
							},
							&cli.IntFlag{
								Name:    "to",
								Aliases: []string{"t"},
								Value:   0,
							},
						},
						Before: func(context *cli.Context) error {
							return loadFid2CidsFromTestFile(
								context.String("test_result_file"),
								context.Int("from"),
								context.Int("to"),
							)
						},
						Subcommands: []*cli.Command{
							{
								// curl -X POST "http://127.0.0.1:5001/api/v0/dht/findprovs?arg=<key>&verbose=<value>&num-providers=20"
								Name: "dht_findprovs",
								Flags: []cli.Flag{
									&cli.BoolFlag{
										Name:    "verbose",
										Aliases: []string{"v"},
										Value:   true,
									},
								},
								Action: func(context *cli.Context) error {
									var input IpfsDhtFindprovsInput
									input.Host = host
									input.Port = port
									input.Method = http.MethodPost
									input.Path = "/api/v0/dht/findprovs"
									input.DropHttpResp = false
									input.TestFile = context.String("test_result_file")
									input.From = context.Int("from")
									input.To = context.Int("to")
									input.Verbose = context.Bool("verbose")
									if e := input.check(); e != nil {
										return e
									}
									iInput = input
									return doIterHttpRequest(input)
								},
							},
							{
								// curl -X POST "http://127.0.0.1:5001/api/v0/dag/stat?arg=<root>&progress=true"
								Name: "dag_stat",
								Flags: []cli.Flag{
									&cli.BoolFlag{
										Name:    "progress",
										Aliases: []string{"v"},
										Value:   true,
									},
								},
								Action: func(context *cli.Context) error {
									var input IpfsDhtFindprovsInput
									input.Host = host
									input.Port = port
									input.Method = http.MethodPost
									input.Path = "/api/v0/dag/stat"
									input.DropHttpResp = false
									input.TestFile = context.String("test_result_file")
									input.From = context.Int("from")
									input.To = context.Int("to")
									input.Verbose = context.Bool("verbose")
									if e := input.check(); e != nil {
										return e
									}
									iInput = input
									return doIterHttpRequest(input)
								},
							},
							{
								// curl -X POST "http://127.0.0.1:5001/api/v0/cat?arg=<ipfs-path>&offset=<value>&length=<value>&progress=true"
								Name: "cat",
								Flags: []cli.Flag{
									&cli.IntFlag{
										Name:    "offset",
										Aliases: []string{"o"},
										Value:   0,
									},
									&cli.IntFlag{
										Name:    "length",
										Aliases: []string{"l"},
										Value:   0, // TODO check api
									},
									&cli.BoolFlag{
										Name:    "progress",
										Aliases: []string{"v"},
										Value:   true,
									},
								},
								Action: func(context *cli.Context) error {
									var input IpfsDhtFindprovsInput
									input.Host = host
									input.Port = port
									input.Method = http.MethodPost
									input.Path = "/api/v0/cat"
									input.DropHttpResp = true
									input.TestFile = context.String("test_result_file")
									input.From = context.Int("from")
									input.To = context.Int("to")
									input.Verbose = context.Bool("verbose")
									if e := input.check(); e != nil {
										return e
									}
									iInput = input
									return doIterHttpRequest(input)
								},
							},
						},
					},
				},
			},
			{
				Name: "gen_file",
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
					input := GenFileInput{
						From: context.Int("from"),
						To:   context.Int("to"),
						Size: context.Int("size"),
					}
					if e := input.check(); e != nil {
						return e
					}
					iInput = input
					return genFiles(input)
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		logger.Fatal(err.Error())
	}
}
