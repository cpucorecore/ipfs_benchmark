package main

import (
	"fmt"
	"os"
	"path"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

const (
	DeviceURandom = "/dev/urandom"
)

const (
	ErrCreateFile = ErrCategoryFile + 1
	ErrReadFile   = ErrCategoryFile + 2
	ErrWriteFile  = ErrCategoryFile + 3
)

func genFiles(input GenFileParams) error {
	chFids := make(chan int, 10000)
	go func() {
		for i := input.From; i < input.To; i++ {
			chFids <- i
		}
		close(chFids)
	}()

	var countResultsWg sync.WaitGroup
	countResultsWg.Add(1)
	go countResults(&countResultsWg)

	var wg sync.WaitGroup
	wg.Add(p.Goroutines)
	for i := 0; i < p.Goroutines; i++ {
		go func() {
			defer wg.Done()

			rf, e := os.Open(DeviceURandom)
			if e != nil {
				logger.Error("open file err", zap.String("file", DeviceURandom), zap.String("err", e.Error()))
				return
			}
			defer rf.Close()

			buffer := make([]byte, 1024*1024)
			for {
				fid, ok := <-chFids
				if !ok {
					return
				}

				r := Result{
					Fid: fid,
				}

				fp := path.Join(PathFiles, fmt.Sprintf("%d", fid))
				fd, e := os.Create(fp)
				if e != nil {
					logger.Error("create file err", zap.String("file", fp), zap.String("err", e.Error()))

					r.Ret = ErrCreateFile
					r.Err = e
					chResults <- r

					continue
				}

				var currentSize, rn, wn int

				if p.SyncConcurrency {
					atomic.AddInt32(&concurrency, 1)
					r.Concurrency = concurrency
				} else {
					r.Concurrency = int32(p.Goroutines)
				}

				r.S = time.Now()
				for currentSize < input.Size {
					rn, e = rf.Read(buffer[:])
					if e != nil {
						logger.Error("read file err", zap.String("err", e.Error()))

						r.Ret = ErrReadFile
						r.Err = e
						chResults <- r
						fd.Close()

						break
					}

					wn, e = fd.Write(buffer[:rn])
					if e != nil {
						logger.Error("write file err", zap.String("err", e.Error()))

						r.Ret = ErrWriteFile
						r.Err = e
						chResults <- r
						fd.Close()

						break
					}

					currentSize += wn
				}
				fd.Close()
				r.E = time.Now()

				if p.SyncConcurrency {
					atomic.AddInt32(&concurrency, -1)
				}

				r.Latency = r.E.Sub(r.S).Microseconds()

				chResults <- r
				if p.Verbose {
					logger.Debug("file generated", zap.String("file", fp))
				}
			}
		}()
	}
	wg.Wait()

	close(chResults)

	countResultsWg.Wait()
	return nil
}
