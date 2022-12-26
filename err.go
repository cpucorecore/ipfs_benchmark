package main

import "errors"

var ErrCheckFailed = errors.New("check failed")

const (
	ErrCategoryFile        = 100
	ErrCategoryHttpRequest = 200
	ErrCategoryHttp        = 300
	ErrCategoryJson        = 400
)
