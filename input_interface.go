package main

type IInput interface {
	name() string
	info() string
	check() bool
}

type ParamsUrl interface {
	paramsUrl() string
}

type IterParamsUrl interface {
	paramsUrl(it string) string
}
