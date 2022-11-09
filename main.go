package main

import (
	"net/http"
	"os"
	"sync"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

const ReportsDir = "reports"

var input Input
var params Params

var resultsAnalyserWg sync.WaitGroup

var chCids = make(chan string, 100000)
var chResults = make(chan Result, 20000)

var testCase string
var logger *zap.Logger

func main() {
	logger, _ = zap.NewDevelopment()

	app := &cli.App{
		Name: "ipfs_api_benchmark",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "timestamp",
				Usage:       "goroutine number",
				Value:       false,
				Destination: &params.Timestamp,
				Aliases:     []string{"t"},
			},
			&cli.IntFlag{
				Name:        "goroutines",
				Usage:       "goroutine number",
				Value:       4,
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
			&cli.BoolFlag{
				Name:        "verbose",
				Usage:       "Verbose log",
				Value:       false,
				Destination: &params.Verbose,
				Aliases:     []string{"v"},
			},
			&cli.StringFlag{
				Name:        "hostPort",
				Usage:       "api host and port",
				Value:       "192.168.0.85:9094",
				Destination: &input.HostPort,
			},
			&cli.IntFlag{
				Name:        "window",
				Usage:       "analyseResults Window",
				Value:       100,
				Destination: &params.Window,
				Aliases:     []string{"w"},
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
			return nil
		},
		Commands: []*cli.Command{
			{
				Name: "compare",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "sort",
						Usage: "sort the values",
						Value: false,
					},
					&cli.StringFlag{
						Name:    "tag",
						Usage:   "tag for file name, final name is compare_${tag}_smooth${smooth}.png",
						Value:   "tag",
						Aliases: []string{"t"},
					},
				},
				Action: func(context *cli.Context) error {
					return Compare(
						context.String("tag"),
						context.Int("smooth"),
						context.Bool("sort"),
						context.Int("window"),
						context.Args().Slice()...)
				},
			},
			{
				Name: "gc",
				Action: func(context *cli.Context) error {
					return gc()
				},
			},
			{
				Name: "cluster",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "Results",
						Usage:       "the file To save Results",
						Value:       "./Results.json",
						Destination: &params.ResultsFile,
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

							t, e := loadTest(params.ResultsFile)
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
									testCase = "ClusterPinGet"
									return doRequests(http.MethodGet, "/pins/")
								},
							},
							{
								Name: "add",
								Action: func(context *cli.Context) error {
									testCase = "ClusterPin"
									return doRequests(http.MethodPost, "/pins/ipfs/")
								},
							},
							{
								Name: "rm",
								Action: func(context *cli.Context) error {
									testCase = "ClusterUnpin"
									return doRequests(http.MethodDelete, "/pins/ipfs/")
								},
							},
						},
					},
					{
						Name: "add",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:        "dir",
								Usage:       "dir for file To creat at",
								Value:       "./files",
								Destination: &params.Dir,
								Aliases:     []string{"d"},
							},
							&cli.IntFlag{
								Name:        "blockSize",
								Usage:       "block size, max value 1048576(1MB), default 262144(256KB)",
								Destination: &input.BlockSize,
								Value:       1024 * 256,
							},
							&cli.BoolFlag{
								Name:    "pin",
								Value:   true,
								Aliases: []string{"p"},
							},
							&cli.IntFlag{
								Name:        "replicationMin",
								Value:       2,
								Destination: &input.ReplicationMin,
							},
							&cli.IntFlag{
								Name:        "replicationMax",
								Value:       2,
								Destination: &input.ReplicationMax,
							},
						},
						Action: func(context *cli.Context) error {
							testCase = "ClusterAdd"
							return sendFiles(context.Bool("pin"))
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
							testCase = "IpfsStat"
							return doRequests(http.MethodPost, "/api/v0/repo/stat")
						},
					},
				},
			},
			{
				Name: "genFile",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "Dir",
						Usage:   "Dir for file To creat at",
						Value:   "./files",
						Aliases: []string{"d"},
					},
					&cli.IntFlag{
						Name:  "size",
						Usage: "file size",
						Value: 1024 * 1024,
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
					testCase = "GenFile"
					return genFiles(context.String("Dir"), context.Int("size"))
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		logger.Fatal(err.Error())
	}
}
