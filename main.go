package main

import (
	"errors"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"

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

// common params
var (
	iInput IInput

	hosts               cli.StringSlice
	p                   HttpParams
	from, to, repeat    int
	testReport, cidFile string
)

// special params
var (
	sortTps, sortLatency                  bool // compare
	size                                  int  // gen_file
	replica                               int  // cluster add, cluster pin add
	pin                                   bool // cluster add
	fileBufferSize, blockSize             int  // cluster add
	verbose_, streams, latency, direction bool // ipfs swarm peers
	progress                              bool // ipfs dag stat
	offset, length                        int  // ipfs cat
)

func main() {
	go func() {
		http.ListenAndServe("0.0.0.0:6666", nil)
	}()

	app := &cli.App{
		Name: "ipfs_benchmark",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "verbose",
				Value:       false,
				Destination: &p.Verbose,
				Aliases:     []string{"v"},
			},
			&cli.IntFlag{
				Name:        "goroutines",
				Value:       1,
				Destination: &p.Goroutines,
				Aliases:     []string{"g"},
			},
			&cli.BoolFlag{
				Name:        "sync_concurrency",
				Value:       true,
				Destination: &p.SyncConcurrency,
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
				Name: "version",
				Action: func(context *cli.Context) error {
					version()
					return nil
				},
			},
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

							input.Verbose = p.Verbose                 // reuse params
							input.Goroutines = p.Goroutines           // reuse params
							input.SyncConcurrency = p.SyncConcurrency // reuse params
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
								Destination: &p.Tag,
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

							input.Tag = p.Tag // reuse params
							input.From = from
							input.To = to
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
					&cli.StringSliceFlag{
						Name:        "hosts",
						Required:    true,
						Destination: &hosts,
					},
					&cli.StringFlag{
						Name:        "port",
						Destination: &p.Port,
						Aliases:     []string{"p"},
						Required:    true,
					},
					&cli.IntFlag{
						Name:        "do_http_timeout",
						Usage:       "http request timeout in second",
						Value:       600,
						Destination: &p.DoHttpTimeout,
						Aliases:     []string{"dto"},
					},
					&cli.IntFlag{
						Name:        "read_http_resp_timeout",
						Usage:       "read http response timeout in second",
						Value:       600,
						Destination: &p.ReadHttpRespTimeout,
						Aliases:     []string{"rto"},
					},
					&cli.IntFlag{
						Name:        "max_retry",
						Value:       3,
						Destination: &p.MaxRetry,
					},
					&cli.BoolFlag{
						Name:        "drop_http_resp",
						Value:       false,
						Destination: &p.DropHttpResp,
						Aliases:     []string{"d"},
					},
					&cli.StringFlag{
						Name:        "tag",
						Usage:       "[crdt/raft], [repo_size-100G]",
						Destination: &p.Tag,
						Required:    true,
					},
				},
				Before: func(context *cli.Context) error {
					httpClient.Timeout = time.Second * time.Duration(p.DoHttpTimeout)

					p.Hosts = hosts.Value()
					if len(p.Hosts) == 0 {
						return errors.New("hosts empty")
					}
					return nil
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
									},
									&cli.StringFlag{
										Name:        "cid_file",
										Destination: &cidFile,
										Aliases:     []string{"c"},
									},
								},
								Before: func(context *cli.Context) error {
									if context.IsSet("test_report") {
										return loadFid2CidsFromTestReport()
									} else if context.IsSet("cid_file") {
										return loadCidFile()
									} else {
										return errors.New("test_report or cid_file not set")
									}
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
											p.Method = http.MethodPost
											p.Path = "/pins/ipfs"

											var input ClusterPinAddInput
											input.HttpParams = p
											input.TestReport = testReport
											input.CidFile = cidFile
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
											p.Method = http.MethodDelete
											p.Path = "/pins/ipfs"

											var input ClusterPinRmInput
											input.HttpParams = p
											input.TestReport = testReport
											input.CidFile = cidFile
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
											p.Method = http.MethodGet
											p.Path = "/pins"

											var input ClusterPinGetInput
											input.HttpParams = p
											input.TestReport = testReport
											input.CidFile = cidFile
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
										Name:        "file_buffer_size",
										Usage:       "file buffer size by MB",
										Destination: &fileBufferSize,
										Value:       11,
										Aliases:     []string{"fbs"},
									},
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
									p.Method = http.MethodPost
									p.Path = "/add"

									var input ClusterAddInput
									input.HttpParams = p
									input.From = from
									input.To = to
									input.FileBufferSize = fileBufferSize
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
											p.Method = http.MethodPost
											p.Path = "/api/v0/swarm/peers"

											var input IpfsSwarmPeersInput
											input.HttpParams = p
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
											p.Method = http.MethodPost
											p.Path = "/api/v0/id"

											var input IpfsIdInput
											input.HttpParams = p
											input.Repeat = repeat

											if !input.check() {
												return ErrCheckFailed
											}
											iInput = input

											return doRepeatHttpInput(input)
										},
									},
									{
										// curl -X POST "http://192.168.0.85:5001/api/v0/repo/stat?size-only=&human="
										Name: "repo_stat",
										Action: func(context *cli.Context) error {
											p.Method = http.MethodPost
											p.Path = "/api/v0/repo/stat"

											var input IpfsRepoStat
											input.HttpParams = p
											input.Repeat = repeat
											input.SizeOnly = true
											input.Human = true

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
									},
									&cli.StringFlag{
										Name:        "cid_file",
										Destination: &cidFile,
										Aliases:     []string{"c"},
									},
								},
								Before: func(context *cli.Context) error {
									if context.IsSet("test_report") {
										return loadFid2CidsFromTestReport()
									} else if context.IsSet("cid_file") {
										return loadCidFile()
									} else {
										return errors.New("test_report or cid_file not set")
									}
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
											p.Method = http.MethodPost
											p.Path = "/api/v0/dht/findprovs"

											var input IpfsDhtFindprovsInput
											input.HttpParams = p
											input.TestReport = testReport
											input.CidFile = cidFile
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
											p.Method = http.MethodPost
											p.Path = "/api/v0/dag/stat"

											var input IpfsDagStatInput
											input.HttpParams = p
											input.TestReport = testReport
											input.CidFile = cidFile
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
											//&cli.IntFlag{
											//	Name:    "length",
											//	Aliases: []string{"l"},
											//	Value:   0, // TODO check api
											//},
											&cli.BoolFlag{
												Name:    "progress",
												Aliases: []string{"prg"},
												Value:   true,
											},
										},
										Action: func(context *cli.Context) error {
											p.Method = http.MethodPost
											p.Path = "/api/v0/cat"

											var input IpfsCatInput
											input.HttpParams = p
											input.TestReport = testReport
											input.CidFile = cidFile
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
