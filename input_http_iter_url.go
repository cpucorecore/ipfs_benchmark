package main

type IterUrlHttpInput IterHttpInput

func (i IterUrlHttpInput) iterUrl(it string) string {
	return "/" + it
}
