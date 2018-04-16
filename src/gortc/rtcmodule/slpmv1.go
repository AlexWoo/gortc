// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//
// RTC SLP manager V1

package rtcmodule

import (
	"gortc/apimodule"
	"net/http"
)

type SLPM_V1 struct {
}

func Slpmv1() apimodule.API {
	return &SLPM_V1{}
}

func (api *SLPM_V1) Get(req *http.Request, paras string) (int,
	*map[string]string, interface{}, *map[int]apimodule.RespCode) {

	return -1, nil, listSLP(), nil
}

func (api *SLPM_V1) Post(req *http.Request, paras string) (int,
	*map[string]string, interface{}, *map[int]apimodule.RespCode) {

	apiname := paras
	filename := req.URL.Query().Get("file")

	return -1, nil, addSLP(apiname, filename), nil
}

func (api *SLPM_V1) Delete(req *http.Request, paras string) (int,
	*map[string]string, interface{}, *map[int]apimodule.RespCode) {

	return -1, nil, delSLP(paras), nil
}
