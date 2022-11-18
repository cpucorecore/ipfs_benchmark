package main

import (
	"errors"
	"fmt"
)

func checkFromTo(from, to int) error {
	if from < 0 || to < 0 || to < from {
		return errors.New(fmt.Sprintf("wrong [from, to), from:%d, to:%d", from, to))
	}
	return nil
}

func fromToParamsStr(from, to int) string {
	return fmt.Sprintf("from-%d_to-%d", from, to)
}
