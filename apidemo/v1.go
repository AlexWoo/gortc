package main

import (
	"net/http"
	"rtclib"
)

type V1 struct {
}

func APIInstance() rtclib.API {
	return &V1{}
}

func (i *V1) Get(req *http.Request, paras string) (int,
	*map[string]string, interface{}, *map[int]rtclib.RespCode) {

	return 0, nil, "Hello", nil
}

func (i *V1) Post(req *http.Request, paras string) (int,
	*map[string]string, interface{}, *map[int]rtclib.RespCode) {

	return 2, nil, "Hello", nil
}

func (i *V1) Delete(req *http.Request, paras string) (int,
	*map[string]string, interface{}, *map[int]rtclib.RespCode) {

	return 2, nil, nil, nil
}
