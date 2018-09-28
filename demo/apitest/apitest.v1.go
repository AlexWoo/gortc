package main

import (
	"fmt"
	"net/http"
	"rtclib"
	"time"
)

type apitestv1 struct {
}

func APIInstance() rtclib.API {
	return &apitestv1{}
}

func (i *apitestv1) Get(req *http.Request, paras string) (int,
	*map[string]string, interface{}, *map[int]rtclib.RespCode) {

	apireq := rtclib.NewAPIRequest("GET",
		"http://101.200.241.232:6200/test/v1/xxx",
		map[string]string{}, nil, 10*time.Second)

	res, err := apireq.Do()
	if err != nil {
		return 4, nil, fmt.Sprintln(err), nil
	}

	ret := map[string]interface{}{
		"status": res.StatusCode(),
		"code":   res.APICode(),
		"body":   res.Body(),
	}

	return 0, nil, ret, nil
}

func (i *apitestv1) Post(req *http.Request, paras string) (int,
	*map[string]string, interface{}, *map[int]rtclib.RespCode) {

	return 3, nil, nil, nil
}

func (i *apitestv1) Delete(req *http.Request, paras string) (int,
	*map[string]string, interface{}, *map[int]rtclib.RespCode) {

	return 3, nil, nil, nil
}
