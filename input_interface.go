package main

type IInput interface {
	name() string
	check() bool
	paramsStr() string
}

type ISampleInput interface {
	IInput
	setSample(int)
	getSample() int
}

type IHttpInput interface {
	ISampleInput
	host() string
	port() string
	method() string
	path() string
	baseUrl() string // "http://" + host() + ":" + port() + path()
	dropHttpRespBody() bool
}

// IConstHttpInput
// url: baseUrl() + urlParams()
type IConstHttpInput interface {
	IHttpInput
	urlParams() string
	repeat() int
}

// IIterUrlHttpInput
// final url = baseUrl() + iterUrl(it) + urlParams()
type IIterUrlHttpInput interface {
	IHttpInput
	iterUrl(it string) string
	urlParams() string
}

// IIterParamsHttpInput
// url = baseUrl() + iterParams(it)
type IIterParamsHttpInput interface {
	IHttpInput
	iterParams(it string) string
}
