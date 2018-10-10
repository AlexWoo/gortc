// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//
// RTC SLP manager V1

package main

import (
	"net/http"
	"rtclib"
)

type SLPM_V1 struct {
}

func Slpmv1() rtclib.API {
	return &SLPM_V1{}
}

func (api *SLPM_V1) Get(req *http.Request, paras string) (int,
	*map[string]string, interface{}, *map[int]rtclib.RespCode) {

	switch paras {
	case "slps":
		return -1, nil, sm.listSLP(), nil
	}

	return 3, nil, nil, nil
}

func (api *SLPM_V1) Post(req *http.Request, paras string) (int,
	*map[string]string, interface{}, *map[int]rtclib.RespCode) {

	apiname := paras
	filename := req.URL.Query().Get("file")

	return -1, nil, sm.addSLP(apiname, filename), nil
}

func (api *SLPM_V1) Delete(req *http.Request, paras string) (int,
	*map[string]string, interface{}, *map[int]rtclib.RespCode) {

	return -1, nil, sm.delSLP(paras), nil
}
