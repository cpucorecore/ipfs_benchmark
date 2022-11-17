package main

var _ IConstHttpInput = ConstHttpInput{}

type ConstHttpInput struct {
	HttpInput
	Repeat int
}

func (i ConstHttpInput) urlParams() string {
	if pf, ok := repetitive[i.Path]; ok {
		return pf()
	}
	return ""
}

func (i ConstHttpInput) repeat() int {
	return i.Repeat
}
