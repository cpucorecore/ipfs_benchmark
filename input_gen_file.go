package main

import "fmt"

var _ LocalInput = GenFileInput{}

type GenFileInput struct {
	BaseInput
	From int
	To   int
	Size int
}

func (i GenFileInput) name() string {
	return "gen_file"
}

func (i GenFileInput) check() bool {
	return i.BaseInput.check() && fromToCheck(i.From, i.To)
}

func (i GenFileInput) paramsStr() string {
	return fmt.Sprintf("%s_%s_size-%d", i.BaseInput.paramsStr(), fromToParamsStr(i.From, i.To), i.Size)
}
