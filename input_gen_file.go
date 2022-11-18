package main

import "fmt"

var _ ISampleInput = GenFileInput{}

type GenFileInput struct {
	SampleInput
	From int
	To   int
	Size int
}

func (i GenFileInput) name() string {
	return "gen_file"
}

func (i GenFileInput) check() bool {
	return fromToCheck(i.From, i.To)
}

func (i GenFileInput) paramsStr() string {
	return fmt.Sprintf("%s_%s_size-%d", baseParamsStr(), fromToParamsStr(i.From, i.To), i.Size)
}
