package main

import "errors"

var ErrCheckFailed = errors.New("check failed")

const (
	ErrCategoryFile = 100
	ErrCategoryHttp = 200
	ErrCategoryJson = 300
)
