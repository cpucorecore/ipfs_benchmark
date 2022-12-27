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
	name := iInput.name() + "_" + iInput.info()

	var ers []ErrResult
	for _, r := range rs.Results {
		if r.Ret != 0 {
			ers = append(ers, ErrResult{R: r, ErrMsg: r.Err.Error()})
		}
	}

	if len(ers) > 0 {
		e := saveErrResults(filepath.Join(ErrsDir, name+JsonSuffix), ers)
		if e != nil {
			logger.Error("saveErrResults err", zap.String("err", e.Error()))
		}
	}

	t := Test{
		Input:          iInput,
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

func adjustTo(itemLen int) bool {
	if to == 0 {
		to = itemLen
	} else {
		if to > itemLen {
			to = itemLen
			if from >= to {
				return false
			}
		}
	}

	return true
}

func loadFid2CidsFromTestReport() error {
	t, e := loadTest(testReport)
	if e != nil {
		logger.Error("loadTest err", zap.String("err", e.Error()))
		return e
	}

	fid2Cids := make([]Fid2Cid, 0, len(t.ResultsSummary.Results))
	for _, r := range t.ResultsSummary.Results {
		if r.Cid != "" {
			fid2Cids = append(fid2Cids, Fid2Cid{Fid: r.Fid, Cid: r.Cid})
		}
	}

	ok := adjustTo(len(fid2Cids))
	if !ok {
		logger.Warn("no valid items by [from, to)")
		close(chFid2Cids)
		return nil
	}

	go func() {
		for _, fie2Cid := range fid2Cids[from:to] {
			chFid2Cids <- fie2Cid
		}
		close(chFid2Cids)
	}()

	return nil
}

func loadCidFile() error {
	bs, e := loadFile(cidFile)
	if e != nil {
		return e
	}

	cids := strings.Split(strings.TrimSpace(string(bs)), "\n")

	ok := adjustTo(len(cids))
	if !ok {
		logger.Warn("no valid items by [from, to)")
		close(chFid2Cids)
		return nil
	}

	go func() {
		for _, cid := range cids[from:to] {
			if cid != "" {
				chFid2Cids <- Fid2Cid{Fid: FakeFid, Cid: cid}
			}
		}
		close(chFid2Cids)
	}()

	return nil
}
