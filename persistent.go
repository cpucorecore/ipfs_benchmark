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

type Test struct {
	Input          IInput
	ResultsSummary ResultsSummary
}

type TestForLoad struct {
	ResultsSummary ResultsSummary
}

func saveTest(rs ResultsSummary) {
	name := ipt.name() + "_" + ipt.paramsStr()

	var ers []ErrResult
	for _, r := range rs.Results {
		if r.Ret != 0 {
			ers = append(ers, ErrResult{R: r, Err: r.Err})
		}
	}

	if len(ers) > 0 {
		e := saveErrResults(filepath.Join(ErrsDir, name+JsonSuffix), ers)
		if e != nil {
			logger.Error("saveErrResults err", zap.String("err", e.Error()))
		}
	}

	t := Test{
		Input:          ipt,
		ResultsSummary: rs,
	}

	bs, e := json.MarshalIndent(t, "", "  ")
	if e != nil {
		logger.Error("MarshalIndent err", zap.String("err", e.Error()))
		return
	}

	e = saveFile(filepath.Join(ReportsDir, name+JsonSuffix), bs)
	if e != nil {
		logger.Error("saveFile err", zap.String("err", e.Error()))
	}
}

func loadTest(name string) (TestForLoad, error) {
	var t TestForLoad

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

type Fid2Cid struct {
	Fid int
	Cid string
}

func loadFid2CidsFromTestFile(testFile string) error {
	t, e := loadTest(testFile)
	if e != nil {
		logger.Error("loadTest err", zap.String("err", e.Error()))
		return e
	}

	//if input.To > len(t.ResultsSummary.Results) {
	//	input.To = len(t.ResultsSummary.Results)
	//}

	go func() {
		//for _, r := range t.ResultsSummary.Results[input.From:input.To] {
		for _, r := range t.ResultsSummary.Results {
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
