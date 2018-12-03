package main

import (
	"net/http"
	"rtclib"
)

type apiaccessv1 struct {
}

func APIInstance() rtclib.API {
	return &apiaccessv1{}
}

func (i *apiaccessv1) Get(req *http.Request, paras string) (int,
	*map[string]string, interface{}, *map[int]rtclib.RespCode) {

	return 3, nil, nil, nil
}

func (i *apiaccessv1) Post(req *http.Request, paras string) (int,
	*map[string]string, interface{}, *map[int]rtclib.RespCode) {

	// access auth successd
	// return 0, nil, nil, nil

	// access auth failed
	return 3, nil, "just for test", nil
}

func (i *apiaccessv1) Delete(req *http.Request, paras string) (int,
	*map[string]string, interface{}, *map[int]rtclib.RespCode) {

	return 3, nil, nil, nil
}
