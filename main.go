package main

import (
	"net/http"
	"os"
	"sync"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

const (
	ReportsDir = "reports"
	ImagesDir  = "images"
)

var (
	params Params
	input  Input

	resultsAnalyserWg sync.WaitGroup

	chCids    = make(chan string, 100000)
	chResults = make(chan Result, 20000)

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
				Value:       10,
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

			e = os.MkdirAll(ReportsDir, os.ModePerm)
			if e != nil {
				logger.Error("create Dir err", zap.String("err", e.Error()))
				return e
			}

			e = os.MkdirAll(ImagesDir, os.ModePerm)
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
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "test_result_file",
						Usage:       "the file To save test results",
						Destination: &params.TestResultFile,
						Aliases:     []string{"trf"},
					},
				},
				Before: func(context *cli.Context) error {
					resultsAnalyserWg.Add(1)
					go resultsAnalyser()
					return nil
				},
				After: func(context *cli.Context) error {
					resultsAnalyserWg.Wait()
					return nil
				},
				Subcommands: []*cli.Command{
					{
						Name: "pin",
						Before: func(context *cli.Context) error {
							defer close(chCids)

							t, e := loadTest(params.TestResultFile)
							if e != nil {
								logger.Error("loadTestResult err", zap.String("err", e.Error()))
								return e
							}

							if input.To > len(t.Results) {
								input.To = len(t.Results)
							}

							for _, r := range t.Results[input.From:input.To] {
								if r.Cid != "" {
									chCids <- r.Cid
								}
							}

							return nil
						},
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
									input.TestCase = "ClusterPin"
									return doRequests(http.MethodPost, "/pins/ipfs/")
								},
							},
							{
								Name: "rm",
								Action: func(context *cli.Context) error {
									input.TestCase = "ClusterUnpin"
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
				},
			},
			{
				Name: "ipfs",
				Before: func(context *cli.Context) error {
					resultsAnalyserWg.Add(1)
					go resultsAnalyser()
					return nil
				},
				After: func(context *cli.Context) error {
					resultsAnalyserWg.Wait()
					return nil
				},
				Subcommands: []*cli.Command{
					{
						Name: "stat",
						Action: func(context *cli.Context) error {
							input.TestCase = "IpfsStat"
							return doRequests(http.MethodPost, "/api/v0/repo/stat")
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
						Name:    "size",
						Usage:   "file size",
						Value:   1024 * 1024,
						Aliases: []string{"s"},
					},
				},
				Before: func(context *cli.Context) error {
					resultsAnalyserWg.Add(1)
					go resultsAnalyser()
					return nil
				},
				After: func(context *cli.Context) error {
					resultsAnalyserWg.Wait()
					return nil
				},
				Action: func(context *cli.Context) error {
					input.TestCase = "GenFile"
					return genFiles(context.Int("size"))
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		logger.Fatal(err.Error())
	}
}
