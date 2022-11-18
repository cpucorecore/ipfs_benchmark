package main

type HttpInput struct {
	Host         string
	Port         string
	Method       string
	Path         string
	DropHttpResp bool
}

func (i HttpInput) name() string {
	return getNameByHttpMethodAndPath(i.Method, i.Path)
}

func (i HttpInput) method() string {
	return i.Method
}

func (i HttpInput) dropHttpResp() bool {
	return i.DropHttpResp
}

func (i HttpInput) baseUrl() string {
	return "http://" + i.Host + ":" + i.Port + i.Path
}
