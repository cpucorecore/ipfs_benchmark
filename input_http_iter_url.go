package main

type IterUrlHttpInput struct {
	IterHttpInput
}

func (i IterHttpInput) iterUrl(it string) string {
	return "/" + it
}
