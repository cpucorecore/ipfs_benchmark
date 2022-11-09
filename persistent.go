package main

import (
	"encoding/json"
	"os"
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
	Input   Input
	Results []Result
}

func saveTest(t Test, ts bool) error {
	b, e := json.Marshal(t)
	if e != nil {
		return e
	}

	fp, e := os.Create(t.Input.info(ts) + ".json")
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
