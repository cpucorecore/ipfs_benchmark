package main

import "fmt"

func fromToCheck(from, to int) bool {
	return from >= 0 && to >= 0 && to >= from
}

func fromToParamsStr(from, to int) string {
	return fmt.Sprintf("from-%d_to-%d", from, to)
}
