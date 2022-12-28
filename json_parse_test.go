package main

import (
	"github.com/buger/jsonparser"
	"testing"
)

func TestParse(t *testing.T) {
	s := `[{"name":"150","cid":"QmRMqvRpR9omduA5J5WW9ymvBV3RB3DqgHTPkzT8c8AHCw","size":10486389,"allocations":["12D3KooWB6cxrTahCGu4T1vLeJTsSU3fHnWJsrig6bNi8afVEekm","12D3KooWJ7b5LSbZJmRvrgGQVoSyVM6bTQdjtSc6cBpLWoZTQKXH"]}]`
	jsonparser.ArrayEach([]byte(s), func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		t.Log(string(value))
		cid, err := jsonparser.GetString(value, "cid")
		t.Log(err)
		t.Log(cid)
	})
}
