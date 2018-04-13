// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//
// API manager V1

package apimodule

import "net/http"

type APIM_V1 struct {
}

func Apimv1() API {
	return &APIM_V1{}
}

func (api *APIM_V1) Get(req *http.Request, paras string) (int,
	*map[string]string, interface{}, *map[int]RespCode) {

	return -1, nil, listAPI(), nil
}

func (api *APIM_V1) Post(req *http.Request, paras string) (int,
	*map[string]string, interface{}, *map[int]RespCode) {

	apiname := paras
	filename := req.Form.Get("file")

	return -1, nil, addAPI(apiname, filename), nil
}

func (api *APIM_V1) Delete(req *http.Request, paras string) (int,
	*map[string]string, interface{}, *map[int]RespCode) {

	return -1, nil, delAPI(paras), nil
}
