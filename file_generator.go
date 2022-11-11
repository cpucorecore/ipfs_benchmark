package main

import (
	"fmt"
	"os"
	"path"
	"sync"
	"time"

	"go.uber.org/zap"
)

func genFiles() error {
	logger.Info("genFiles",
		zap.String("FilesDir", params.FilesDir),
		zap.Int("From", input.From),
		zap.Int("To", input.To),
		zap.Int("FileSize", params.FileSize),
	)

	e := os.MkdirAll(params.FilesDir, os.ModePerm)
	if e != nil {
		logger.Error("MkdirAll err", zap.String("dir", params.FilesDir), zap.String("err", e.Error()))
		return e
	}

	chFids := make(chan int, 10000)
	go func() {
		for i := input.From; i < input.To; i++ {
			chFids <- i
		}
		close(chFids)
	}()

	var prsWg sync.WaitGroup
	prsWg.Add(1)
	go countResults(&prsWg)

	var wg sync.WaitGroup
	for i := 0; i < input.Goroutines; i++ {
		wg.Add(1)

		go func(gid int) {
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
					Gid: gid,
					Fid: fid,
				}

				fp := path.Join(params.FilesDir, fmt.Sprintf("%d", fid))
				fd, e := os.Create(fp)
				if e != nil {
					logger.Error("create file err", zap.String("file", fp), zap.String("err", e.Error()))

					r.Ret = -1
					r.Error = e
					chResults <- r

					continue
				}

				var currentSize, rn, wn int
				r.S = time.Now()
				for currentSize < params.FileSize {
					rn, e = rf.Read(buffer[:])
					if e != nil {
						logger.Error("read file err", zap.String("err", e.Error()))

						r.Ret = -2
						r.Error = e
						chResults <- r

						break
					}

					wn, e = fd.Write(buffer[:rn])
					if e != nil {
						logger.Error("write file err", zap.String("err", e.Error()))

						r.Ret = -3
						r.Error = e
						chResults <- r

						break
					}

					currentSize += wn
				}
				fd.Close()
				r.E = time.Now()

				chResults <- r
				if params.Verbose {
					logger.Info("file generated", zap.String("file", fp))
				}
			}
		}(i)
	}
	wg.Wait()

	close(chResults)

	prsWg.Wait()
	return nil
}
