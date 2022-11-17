package main

var _ IIterUrlHttpInput = IterUrlHttpInput{}

type IterUrlHttpInput struct {
	HttpInput
	TestFile string
}

func (i IterUrlHttpInput) iterUrl(path string) string {
	return "/" + path
}

func (i IterUrlHttpInput) urlParams() string {
	if pf, ok := repetitive[i.Path]; ok {
		return pf()
	}
	return ""
}
