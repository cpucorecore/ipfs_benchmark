package main

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

func File2bytes(name string) ([]byte, error) {
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

type Test struct {
	Params  Params
	Input   Input
	Results []Result
}

func saveTest(t Test, ts bool) error {
	b, e := json.MarshalIndent(t, "", "  ")
	if e != nil {
		return e
	}

	fp, e := os.Create(filepath.Join(TestResultDir, t.Input.info(ts)+".json"))
	if e != nil {
		return e
	}
	defer fp.Close()

	_, e = fp.Write(b)
	if e != nil {
		return e
	}

	return nil
}

func loadTest(name string) (Test, error) {
	var t Test

	b, e := File2bytes(name)
	if e != nil {
		return t, e
	}

	e = json.Unmarshal(b, &t)
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

	if input.To > len(t.Results) {
		input.To = len(t.Results)
	}

	for _, r := range t.Results[input.From:input.To] {
		if r.Cid != "" {
			chFid2Cids <- Fid2Cid{Fid: r.Fid, Cid: r.Cid}
		}
	}

	return nil
}
