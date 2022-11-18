package main

type RepeatHttpInput struct {
	HttpInput
	Repeat int
}

func (i RepeatHttpInput) repeat() int {
	return i.Repeat
}
