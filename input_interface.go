package main

type IInput interface {
	name() string
	check() error
	paramsStr() string
}

type IHttpInput interface {
	IInput
	method() string
	dropHttpResp() bool
	baseUrl() string
	url(it string) string
}

type IConstHttpInput interface {
	IHttpInput
	urlParams() string
}

type IHttpInputRepeat interface {
	IConstHttpInput
	repeat() int
}

type IHttpInputIterUrl interface {
	IHttpInput
	iterUrl(it string) string
	urlParams() string
}

type IHttpInputIterParams interface {
	IHttpInput
	iterUrlParams(it string) string
}
