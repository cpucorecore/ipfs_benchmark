package main

var _ IIterParamsHttpInput = IterParamsHttpInput{}

type IterParamsHttpInput struct {
	HttpInput
	TestFile string
}

func (i IterParamsHttpInput) iterParams(it string) string {
	if pf, ok := iterative[i.Path]; ok {
		return pf(it)
	}
	return ""
}
