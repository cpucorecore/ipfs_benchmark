package main

import (
	"errors"
	"net/http"
	"os"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
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

var ErrCheckFailed = errors.New("check failed")

var (
	tag                                   string // compare
	sortTps, sortLatency                  bool   // compare
	size                                  int    // gen_file
	verbose                               bool
	goroutines                            int
	syncConcurrency                       bool
	from, to, repeat                      int
	host, port, method, apiPath           string
	dropHttpResp                          bool
	testReport                            string
	replica                               int
	pin                                   bool
	blockSize                             int
	verbose_, streams, latency, direction bool
	progress                              bool
	offset, length                        int
)

var iInput IInput

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
				Value:       true,
				Destination: &syncConcurrency,
				Aliases:     []string{"sc"},
			},
			&cli.IntFlag{
				Name:        "from",
				Destination: &from,
				Aliases:     []string{"f"},
			},
			&cli.IntFlag{
				Name:        "to",
				Destination: &to,
				Aliases:     []string{"t"},
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
						Destination: &host,
						Required:    true,
					},
					&cli.StringFlag{
						Name:        "port",
						Destination: &port,
						Aliases:     []string{"p"},
						Required:    true,
					},
					&cli.BoolFlag{
						Name:        "drop_http_resp",
						Value:       false,
						Destination: &dropHttpResp,
						Aliases:     []string{"d"},
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
								Name: "info",
								Flags: []cli.Flag{
									&cli.BoolFlag{
										Name:    "node_detail",
										Value:   false,
										Aliases: []string{"nd"},
									},
									&cli.BoolFlag{
										Name:    "cid_detail",
										Value:   false,
										Aliases: []string{"cd"},
									},
								},
								Action: func(context *cli.Context) error {
									clusterInfo(context.Bool("node_detail"), context.Bool("cid_detail"), true)
									return nil
								},
							},
							{
								Name: "pin",
								Flags: []cli.Flag{
									&cli.StringFlag{
										Name:        "test_report",
										Destination: &testReport,
										Aliases:     []string{"tr"},
										Required:    true,
									},
								},
								Before: func(context *cli.Context) error {
									return loadFid2CidsFromTestReport()
								},
								Subcommands: []*cli.Command{
									{
										Name: "add",
										Flags: []cli.Flag{
											&cli.IntFlag{
												Name:        "replica",
												Destination: &replica,
												Required:    true,
												Aliases:     []string{"r"},
											},
										},
										Action: func(context *cli.Context) error {
											var input ClusterPinAddInput

											method = http.MethodPost
											apiPath = "/pins/ipfs"

											input.Verbose = verbose
											input.Goroutines = goroutines
											input.SyncConcurrency = syncConcurrency
											input.Host = host
											input.Port = port
											input.Method = method
											input.Path = apiPath
											input.DropHttpResp = dropHttpResp
											input.TestReport = testReport
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

											method = http.MethodDelete
											apiPath = "/pins/ipfs"

											input.Verbose = verbose
											input.Goroutines = goroutines
											input.SyncConcurrency = syncConcurrency
											input.Host = host
											input.Port = port
											input.Method = method
											input.Path = apiPath
											input.DropHttpResp = dropHttpResp
											input.TestReport = testReport
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

											method = http.MethodGet
											apiPath = "/pins"

											input.Verbose = verbose
											input.Goroutines = goroutines
											input.SyncConcurrency = syncConcurrency
											input.Host = host
											input.Port = port
											input.Method = method
											input.Path = apiPath
											input.DropHttpResp = dropHttpResp
											input.TestReport = testReport
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
										Name:        "block_size",
										Usage:       "block size, max value 1048576(1MB)",
										Destination: &blockSize,
										Required:    true,
										Aliases:     []string{"bs"},
									},
									&cli.IntFlag{
										Name:        "replica",
										Destination: &replica,
										Required:    true,
										Aliases:     []string{"r"},
									},
									&cli.BoolFlag{
										Name:        "pin",
										Destination: &pin,
										Required:    true,
										Aliases:     []string{"p"},
									},
								},
								Action: func(context *cli.Context) error {
									var input ClusterAddInput

									method = http.MethodPost
									apiPath = "/add"

									input.Verbose = verbose
									input.Goroutines = goroutines
									input.SyncConcurrency = syncConcurrency
									input.Host = host
									input.Port = port
									input.Method = method
									input.Path = apiPath
									input.DropHttpResp = dropHttpResp
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
										Aliases:  []string{"c"},
										Required: true,
									},
								},
								Before: func(context *cli.Context) error {
									return loadCidFile(context.String("cid_file"))
								},
								Action: func(context *cli.Context) error {
									var input ClusterUnpinByCidInput

									method = http.MethodDelete
									apiPath = "/pins/ipfs"

									input.Verbose = verbose
									input.Goroutines = goroutines
									input.SyncConcurrency = syncConcurrency
									input.Host = host
									input.Port = port
									input.Method = method
									input.Path = apiPath
									input.DropHttpResp = dropHttpResp
									input.cidFile = context.String("cid_file")
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
									&cli.IntFlag{
										Name:        "repeat",
										Destination: &repeat,
										Aliases:     []string{"r"},
										Required:    true,
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
												Aliases:     []string{"vv"},
											},
											&cli.BoolFlag{
												Name:        "streams",
												Destination: &streams,
												Value:       true,
												Aliases:     []string{"s"},
											},
											&cli.BoolFlag{
												Name:        "latency",
												Destination: &latency,
												Value:       true,
												Aliases:     []string{"l"},
											},
											&cli.BoolFlag{
												Name:        "direction",
												Destination: &direction,
												Value:       true,
											},
										},
										Action: func(context *cli.Context) error {
											var input IpfsSwarmPeersInput

											method = http.MethodPost
											apiPath = "/api/v0/swarm/peers"

											input.Verbose = verbose
											input.Goroutines = goroutines
											input.SyncConcurrency = syncConcurrency
											input.Host = host
											input.Port = port
											input.Method = method
											input.Path = apiPath
											input.DropHttpResp = dropHttpResp
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

											method = http.MethodPost
											apiPath = "/api/v0/id"

											input.Verbose = verbose
											input.Goroutines = goroutines
											input.SyncConcurrency = syncConcurrency
											input.Host = host
											input.Port = port
											input.Method = method
											input.Path = apiPath
											input.DropHttpResp = dropHttpResp
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
										Name:        "test_report",
										Destination: &testReport,
										Aliases:     []string{"tr"},
										Required:    true,
									},
								},
								Before: func(context *cli.Context) error {
									return loadFid2CidsFromTestReport()
								},
								Subcommands: []*cli.Command{
									{
										// curl -X POST "http://127.0.0.1:5001/api/v0/dht/findprovs?arg=<key>&Verbose=<value>&num-providers=20"
										Name: "dht_findprovs",
										Flags: []cli.Flag{
											&cli.BoolFlag{
												Name:        "verbose_",
												Destination: &verbose_,
												Value:       true,
												Aliases:     []string{"vv"},
											},
										},
										Action: func(context *cli.Context) error {
											var input IpfsDhtFindprovsInput

											method = http.MethodPost
											apiPath = "/api/v0/dht/findprovs"

											input.Verbose = verbose
											input.Goroutines = goroutines
											input.SyncConcurrency = syncConcurrency
											input.Host = host
											input.Port = port
											input.Method = method
											input.Path = apiPath
											input.DropHttpResp = dropHttpResp
											input.TestReport = testReport
											input.From = from
											input.To = to
											input.Verbose_ = verbose_

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
												Name:        "progress",
												Destination: &progress,
												Aliases:     []string{"p"},
												Value:       true,
											},
										},
										Action: func(context *cli.Context) error {
											var input IpfsDagStatInput

											method = http.MethodPost
											apiPath = "/api/v0/dag/stat"

											input.Verbose = verbose
											input.Goroutines = goroutines
											input.SyncConcurrency = syncConcurrency
											input.Host = host
											input.Port = port
											input.Method = method
											input.Path = apiPath
											input.DropHttpResp = dropHttpResp
											input.TestReport = testReport
											input.From = from
											input.To = to
											input.Progress = progress

											if !input.check() {
												return ErrCheckFailed
											}
											iInput = input

											return doIterParamsRequest(input)
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
												Aliases: []string{"prg"},
												Value:   true,
											},
										},
										Action: func(context *cli.Context) error {
											var input IpfsCatInput

											method = http.MethodPost
											apiPath = "/api/v0/cat"

											input.Verbose = verbose
											input.Goroutines = goroutines
											input.SyncConcurrency = syncConcurrency
											input.Host = host
											input.Port = port
											input.Method = method
											input.Path = apiPath
											input.DropHttpResp = dropHttpResp
											input.TestReport = testReport
											input.From = from
											input.To = to
											input.Offset = offset
											input.Length = length
											input.Progress = progress

											if !input.check() {
												return ErrCheckFailed
											}
											iInput = input

											return doIterParamsRequest(input)
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
