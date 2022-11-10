package main

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

const (
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

func saveTest(rs ResultsSummary) {
	var ers []ErrResult
	for _, r := range rs.Results {
		if r.Ret != 0 {
			ers = append(ers, ErrResult{Result: r, Err: r.Error})
		}
	}

	title := input.info()
	bs, e := json.MarshalIndent(ers, "", "  ")
	if e != nil {
		logger.Error("MarshalIndent err", zap.String("err", e.Error()))
	} else {
		e = saveFile(filepath.Join(ErrsDir, title+JsonSuffix), bs)
		if e != nil {
			logger.Error("saveFile err", zap.String("err", e.Error()))
		}
	}

	t := Test{
		Params:         params,
		Input:          input,
		ResultsSummary: rs,
	}

	bs, e = json.MarshalIndent(t, "", "  ")
	if e != nil {
		logger.Error("MarshalIndent err", zap.String("err", e.Error()))
	} else {
		e = saveFile(filepath.Join(ReportsDir, title+JsonSuffix), bs)
		if e != nil {
			logger.Error("saveFile err", zap.String("err", e.Error()))
		}
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

func loadTestCids(_ *cli.Context) error {
	defer close(chFid2Cids)

	t, e := loadTest(params.TestResultFile)
	if e != nil {
		logger.Error("loadTest err", zap.String("err", e.Error()))
		return e
	}

	if input.To > len(t.ResultsSummary.Results) {
		input.To = len(t.ResultsSummary.Results)
	}

	for _, r := range t.ResultsSummary.Results[input.From:input.To] {
		if r.Cid != "" {
			chFid2Cids <- Fid2Cid{Fid: r.Fid, Cid: r.Cid}
		}
	}

	return nil
}
