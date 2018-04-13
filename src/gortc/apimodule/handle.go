// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//
// API Handle

package apimodule

import (
	"net/http"
	"regexp"
)

func parseUri(uri string) (bool, string, string, string) {
	reg := regexp.MustCompile(`^/(\w+)/(v\d+)/(.+)`)

	match := reg.FindStringSubmatch(uri)
	if len(match) == 0 {
		return false, "", "", ""
	}

	return true, match[1], match[2], match[3]
}

func callAPI(req *http.Request, apiname string, version string,
	paras string) (int, *map[string]string, interface{}, *map[int]RespCode) {

	api := getAPI(apiname + "." + version)
	if api == nil {
		return 3, nil, nil, nil
	}

	switch req.Method {
	case "GET":
		return api.Get(req, paras)
	case "POST":
		return api.Post(req, paras)
	case "DELETE":
		return api.Delete(req, paras)
	}

	return 2, nil, nil, nil
}

func handler(w http.ResponseWriter, req *http.Request) {
	ok, apiname, version, paras := parseUri(req.RequestURI)
	if !ok {
		NewResponse(1, nil, nil, nil).SendResp(w)
		return
	}

	NewResponse(callAPI(req, apiname, version, paras)).SendResp(w)
}
