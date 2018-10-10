// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//
// API manager V1

package main

import (
	"net/http"
	"rtclib"
)

type APIM_V1 struct {
}

func Apimv1() rtclib.API {
	return &APIM_V1{}
}

func (api *APIM_V1) Get(req *http.Request, paras string) (int,
	*map[string]string, interface{}, *map[int]rtclib.RespCode) {

	switch paras {
	case "apis":
		return -1, nil, am.listAPI(), nil
	}

	return 3, nil, nil, nil
}

func (api *APIM_V1) Post(req *http.Request, paras string) (int,
	*map[string]string, interface{}, *map[int]rtclib.RespCode) {

	apiname := paras
	filename := req.URL.Query().Get("file")

	return -1, nil, am.addAPI(apiname, filename), nil
}

func (api *APIM_V1) Delete(req *http.Request, paras string) (int,
	*map[string]string, interface{}, *map[int]rtclib.RespCode) {

	return -1, nil, am.delAPI(paras), nil
}
