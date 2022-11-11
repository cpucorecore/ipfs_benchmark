package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
)

const (
	FakeFid    = -1
	JsonSuffix = ".json"
)

func loadFile(name string) ([]byte, error) {
	fi, e := os.Stat(name)
	if e != nil {
		return nil, e
	}

	f, e := os.Open(name)
	if e != nil {
		return nil, e
	}
	defer f.Close()

	bs := make([]byte, fi.Size())
	_, e = f.Read(bs)
	if e != nil {
		return nil, e
	}

	return bs, nil
}

func saveFile(name string, bs []byte) error {
	fp, e := os.Create(name)
	if e != nil {
		return e
	}
	defer fp.Close()

	_, e = fp.Write(bs)
	if e != nil {
		return e
	}

	return nil
}

type Test struct {
	Params         Params
	Input          Input
	ResultsSummary ResultsSummary
}

func saveErrResults(file string, ers []ErrResult) error {
	bs, e := json.MarshalIndent(ers, "", "  ")
	if e != nil {
		return e
	} else {
		e = saveFile(file, bs)
		if e != nil {
			return e
		}
	}

	return nil
}

func saveTest(rs ResultsSummary) {
	title := input.info()

	var ers []ErrResult
	for _, r := range rs.Results {
		if r.Ret != 0 {
			ers = append(ers, ErrResult{Result: r, Err: r.Error})
		}
	}

	if len(ers) > 0 {
		e := saveErrResults(filepath.Join(ErrsDir, title+JsonSuffix), ers)
		if e != nil {
			logger.Error("saveErrResults err", zap.String("err", e.Error()))
		}
	}

	t := Test{
		Params:         params,
		Input:          input,
		ResultsSummary: rs,
	}

	bs, e := json.MarshalIndent(t, "", "  ")
	if e != nil {
		logger.Error("MarshalIndent err", zap.String("err", e.Error()))
		return
	}

	e = saveFile(filepath.Join(ReportsDir, title+JsonSuffix), bs)
	if e != nil {
		logger.Error("saveFile err", zap.String("err", e.Error()))
	}
}

func loadTest(name string) (Test, error) {
	var t Test

	bs, e := loadFile(name)
	if e != nil {
		return t, e
	}

	e = json.Unmarshal(bs, &t)
	if e != nil {
		return t, e
	}

	return t, nil
}

func loadTestCids() error {
	t, e := loadTest(params.TestResultFile)
	if e != nil {
		logger.Error("loadTest err", zap.String("err", e.Error()))
		return e
	}

	if input.To > len(t.ResultsSummary.Results) {
		input.To = len(t.ResultsSummary.Results)
	}

	go func() {
		for _, r := range t.ResultsSummary.Results[input.From:input.To] {
			if r.Cid != "" {
				chFid2Cids <- Fid2Cid{Fid: r.Fid, Cid: r.Cid}
			}
		}
		close(chFid2Cids)
	}()

	return nil
}

func loadFileCids(file string) error {
	bs, e := loadFile(file)
	if e != nil {
		return e
	}

	cids := strings.Split(strings.TrimSpace(string(bs)), "\n")
	go func() {
		for _, cid := range cids {
			if cid != "" {
				chFid2Cids <- Fid2Cid{Fid: FakeFid, Cid: cid}
			}
		}
		close(chFid2Cids)
	}()

	return nil
}
