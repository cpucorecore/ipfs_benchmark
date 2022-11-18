package main

import (
	"errors"
	"fmt"
	"net/http"
)

var _ IHttpInput = HttpInput{}

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

func (i HttpInput) check() error {
	if i.Method != http.MethodGet && i.Method != http.MethodPost && i.Method != http.MethodDelete {
		return errors.New(fmt.Sprintf("wrong http method:(%s), support:[POST,DELETE,GET]", i.Method))
	}
	if len(i.Path) == 0 || i.Path[0] != '/' {
		return errors.New(fmt.Sprintf("wrong path:(%s)", i.Path))
	}
	return nil
}

func (i HttpInput) paramsStr() string {
	return baseParamsStr() // TODO check
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

func (i HttpInput) url(_ string) string { // TODO check useless
	return i.baseUrl()
}
