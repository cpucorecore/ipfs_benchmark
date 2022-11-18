package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

var (
	iInput IInput

	chFid2Cids = make(chan Fid2Cid, 20000)
	chResults  = make(chan Result, 30000)

	logger *zap.Logger
)

func baseParamsStr() string {
	return fmt.Sprintf("v-%v_g-%d_s-%v", verbose, goroutines, syncConcurrency)
}

func init() {
	logger, _ = zap.NewDevelopment()

	ec := initDirs()
	if ec > 0 {
		logger.Error("initDirs failed", zap.Int("failed", ec))
		os.Exit(-1)
	}
}

var ErrCheckFailed = errors.New("check failed")

var (
	tag                                   string // compare
	sortTps, sortLatency                  bool   // compare
	size                                  int    // gen_file
	verbose                               bool
	goroutines                            int
	syncConcurrency                       bool
	from, to                              int
	host, port, method, path              string
	dropHttpResp                          bool
	repeat                                int
	testFile                              string
	replica                               int
	pin                                   bool
	blockSize                             int
	verbose_, streams, latency, direction bool
)

func main() {
	app := &cli.App{
		Name: "ipfs_benchmark",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "verbose",
				Value:       false,
				Destination: &verbose,
				Aliases:     []string{"v"},
			},
			&cli.IntFlag{
				Name:        "goroutines",
				Value:       1,
				Destination: &goroutines,
				Aliases:     []string{"g"},
			},
			&cli.BoolFlag{
				Name:        "sync_concurrency",
				Value:       false,
				Destination: &syncConcurrency,
				Aliases:     []string{"s"},
			},
		},
		Commands: []*cli.Command{
			{
				Name: "tool",
				Subcommands: []*cli.Command{
					{
						Name: "gen_file",
						Flags: []cli.Flag{
							&cli.IntFlag{
								Name:        "from",
								Value:       0,
								Destination: &from,
								Aliases:     []string{"f"},
							},
							&cli.IntFlag{
								Name:        "to",
								Value:       10,
								Destination: &to,
								Aliases:     []string{"t"},
							},
							&cli.IntFlag{
								Name:        "size",
								Value:       1024 * 1024,
								Destination: &size,
								Aliases:     []string{"s"},
							},
						},
						Action: func(context *cli.Context) error {
							var input GenFileParams

							input.Verbose = verbose
							input.Goroutines = goroutines
							input.SyncConcurrency = syncConcurrency
							input.From = from
							input.To = to
							input.Size = size

							if !input.check() {
								return ErrCheckFailed
							}
							iInput = input

							return genFiles(input)
						},
					},
					{
						Name: "compare",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:        "tag", // TODO remove this flag, instead by infos from [test files...]
								Aliases:     []string{"t"},
								Destination: &tag,
								Required:    true,
							},
							&cli.BoolFlag{
								Name:        "sort_tps",
								Value:       true,
								Destination: &sortTps,
								Aliases:     []string{"st"},
							},
							&cli.BoolFlag{
								Name:        "sort_latency",
								Value:       true,
								Destination: &sortLatency,
								Aliases:     []string{"sl"},
							},
						},
						Action: func(context *cli.Context) error {
							var input CompareParams

							input.Tag = tag
							input.SortTps = sortTps
							input.SortLatency = sortLatency

							if !input.check() {
								return ErrCheckFailed
							}
							iInput = input

							return CompareTests(input, context.Args().Slice()...)
						},
					},
				},
			},
			{
				Name: "api",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "host",
						Value:       "127.0.0.1",
						Destination: &host,
						Required:    true,
					},
					&cli.StringFlag{
						Name:        "port",
						Value:       "9094",
						Destination: &port,
						Aliases:     []string{"p"},
						Required:    true,
					},
				},
				Subcommands: []*cli.Command{
					{
						Name: "cluster",
						Subcommands: []*cli.Command{
							{
								Name: "gc",
								Action: func(context *cli.Context) error {
									return gc()
								},
							},
							{
								Name: "pin",
								Flags: []cli.Flag{
									&cli.StringFlag{
										Name:        "test_file",
										Destination: &testFile,
										Aliases:     []string{"tf"},
										Required:    true,
									},
									&cli.IntFlag{
										Name:        "from",
										Value:       0,
										Destination: &from,
										Aliases:     []string{"f"},
									},
									&cli.IntFlag{
										Name:        "to",
										Value:       10,
										Destination: &to,
										Aliases:     []string{"t"},
									},
								},
								Before: func(context *cli.Context) error {
									return loadFid2CidsFromTestFile()
								},
								Subcommands: []*cli.Command{
									{
										Name: "add",
										Flags: []cli.Flag{
											&cli.IntFlag{
												Name:        "replica",
												Value:       2,
												Destination: &replica,
											},
										},
										Action: func(context *cli.Context) error {
											var input ClusterPinAddInput

											input.Verbose = verbose
											input.Goroutines = goroutines
											input.SyncConcurrency = syncConcurrency
											input.Host = host
											input.Port = port
											input.Method = http.MethodPost
											input.Path = "/pins/ipfs"
											input.DropHttpResp = false
											input.TestFile = testFile
											input.From = from
											input.To = to
											input.Replica = replica

											if !input.check() {
												return ErrCheckFailed
											}
											iInput = input

											return doIterUrlRequest(input)
										},
									},
									{
										Name: "rm",
										Action: func(context *cli.Context) error {
											var input ClusterPinRmInput

											input.Verbose = verbose
											input.Goroutines = goroutines
											input.SyncConcurrency = syncConcurrency
											input.Host = host
											input.Port = port
											input.Method = http.MethodDelete
											input.Path = "/pins/ipfs"
											input.DropHttpResp = false
											input.TestFile = testFile
											input.From = from
											input.To = to

											if !input.check() {
												return ErrCheckFailed
											}
											iInput = input

											return doIterUrlRequest(input)
										},
									},
									{
										Name: "get",
										Action: func(context *cli.Context) error {
											var input ClusterPinGetInput

											input.Verbose = verbose
											input.Goroutines = goroutines
											input.SyncConcurrency = syncConcurrency
											input.Host = host
											input.Port = port
											input.Method = http.MethodGet
											input.Path = "/pins"
											input.DropHttpResp = false
											input.TestFile = testFile
											input.From = from
											input.To = to

											if !input.check() {
												return ErrCheckFailed
											}
											iInput = input

											return doIterUrlRequest(input)
										},
									},
								},
							},
							{
								Name: "add",
								Flags: []cli.Flag{
									&cli.IntFlag{
										Name:        "from",
										Value:       0,
										Destination: &from,
										Aliases:     []string{"f"},
									},
									&cli.IntFlag{
										Name:        "to",
										Value:       10,
										Destination: &to,
										Aliases:     []string{"t"},
									},
									&cli.IntFlag{
										Name:        "block_size",
										Usage:       "block size, max value 1048576(1MB), default 1MB",
										Value:       1024 * 1024,
										Destination: &blockSize,
										Aliases:     []string{"bs"},
									},
									&cli.IntFlag{
										Name:        "replica",
										Value:       2,
										Destination: &replica,
										Aliases:     []string{"r"},
									},
									&cli.BoolFlag{
										Name:        "pin",
										Value:       true,
										Destination: &pin,
										Aliases:     []string{"p"},
									},
								},
								Action: func(context *cli.Context) error {
									var input ClusterAddInput

									input.Verbose = verbose
									input.Goroutines = goroutines
									input.SyncConcurrency = syncConcurrency
									input.Host = host
									input.Port = port
									input.Method = http.MethodPost
									input.Path = "/add"
									input.DropHttpResp = false
									input.TestFile = testFile
									input.From = from
									input.To = to
									input.BlockSize = blockSize
									input.Replica = replica
									input.Pin = pin

									if !input.check() {
										return ErrCheckFailed
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

									input.Verbose = verbose
									input.Goroutines = goroutines
									input.SyncConcurrency = syncConcurrency
									input.Host = host
									input.Port = port
									input.Method = http.MethodDelete
									input.Path = "/pins/ipfs"
									input.DropHttpResp = false
									input.TestFile = testFile
									input.From = from
									input.To = to

									if !input.check() {
										return ErrCheckFailed
									}
									iInput = input

									return doIterUrlRequest(input)
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
												Name:        "verbose_",
												Destination: &verbose_,
												Value:       true,
											},
											&cli.BoolFlag{
												Name:        "streams",
												Destination: &streams,
												Value:       true,
											},
											&cli.BoolFlag{
												Name:        "latency",
												Destination: &latency,
												Value:       true,
											},
											&cli.BoolFlag{
												Name:        "direction",
												Destination: &direction,
												Value:       true,
											},
										},
										Action: func(context *cli.Context) error {
											var input IpfsSwarmPeersInput

											input.Verbose = verbose
											input.Goroutines = goroutines
											input.SyncConcurrency = syncConcurrency
											input.Host = host
											input.Port = port
											input.Method = http.MethodPost
											input.Path = "/api/v0/swarm/peers"
											input.DropHttpResp = false
											input.Repeat = repeat
											input.Verbose_ = verbose_
											input.Streams = streams
											input.Latency = latency
											input.Direction = direction

											if !input.check() {
												return ErrCheckFailed
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

											input.Verbose = verbose
											input.Goroutines = goroutines
											input.SyncConcurrency = syncConcurrency
											input.Host = host
											input.Port = port
											input.Method = http.MethodPost
											input.Path = "/api/v0/id"
											input.DropHttpResp = false
											input.Repeat = repeat

											if !input.check() {
												return ErrCheckFailed
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
									return loadFid2CidsFromTestFile()
								},
								Subcommands: []*cli.Command{
									{
										// curl -X POST "http://127.0.0.1:5001/api/v0/dht/findprovs?arg=<key>&Verbose=<value>&num-providers=20"
										Name: "dht_findprovs",
										Flags: []cli.Flag{
											&cli.BoolFlag{
												Name:    "Verbose",
												Aliases: []string{"v"},
												Value:   true,
											},
										},
										Action: func(context *cli.Context) error {
											var input IpfsDhtFindprovsInput

											sssssssssssssssssbbb

											input.Host = host
											input.Port = port
											input.Method = http.MethodPost
											input.Path = "/api/v0/dht/findprovs"
											input.DropHttpResp = false
											input.TestFile = context.String("test_result_file")
											input.From = context.Int("from")
											input.To = context.Int("to")
											input.Verbose = context.Bool("Verbose")

											if !input.check() {
												return ErrCheckFailed
											}
											iInput = input

											return doIterParamsRequest(input)
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
											input.Verbose = context.Bool("Verbose")
											if e := input.check(); e != nil {
												return e
											}
											iInput = input
											return doIterUrlRequest(input)
										},
									},
									{
										// curl -X POST "http://127.0.0.1:5001/api/v0/cat?arg=<ipfs-Path>&offset=<value>&length=<value>&progress=true"
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
											input.Verbose = context.Bool("Verbose")
											if e := input.check(); e != nil {
												return e
											}
											iInput = input
											return doIterUrlRequest(input)
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		logger.Fatal(err.Error())
	}
}
