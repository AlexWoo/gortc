// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//
// MainModule

package main

import (
	"net/http"
	"rtclib"
	"runtime"
)

type RUNTIME_V1 struct {
}

func RunTimeV1() rtclib.API {
	return &RUNTIME_V1{}
}

func stack() string {
	l := 4096
	b := make([]byte, l)

	for {
		if l > runtime.Stack(b, true) {
			break
		}

		l *= 2
		b = make([]byte, l)
	}

	return string(b)
}

func (api *RUNTIME_V1) Get(req *http.Request, paras string) (int,
	*map[string]string, interface{}, *map[int]rtclib.RespCode) {

	switch paras {
	case "stack": // GO Stack
		return -1, nil, stack(), nil
	case "jstack": // JSIP Stack
		return -1, nil, rtclib.JStackInstance().State(), nil
	}
	return 3, nil, nil, nil
}

func (api *RUNTIME_V1) Post(req *http.Request, paras string) (int,
	*map[string]string, interface{}, *map[int]rtclib.RespCode) {

	return 2, nil, nil, nil
}

func (api *RUNTIME_V1) Delete(req *http.Request, paras string) (int,
	*map[string]string, interface{}, *map[int]rtclib.RespCode) {

	return 2, nil, nil, nil
}
