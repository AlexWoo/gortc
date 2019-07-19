// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//
// Base Type

package rtclib

import (
	"net/http"
	"strings"
)

// API Reponse Code
type RespCode struct {
	Status int
	Msg    string
}

// API interface
type API interface {
	Get(req *http.Request, paras string) (int,
		*map[string]string, interface{}, *map[int]RespCode)
	Post(req *http.Request, paras string) (int,
		*map[string]string, interface{}, *map[int]RespCode)
	Delete(req *http.Request, paras string) (int,
		*map[string]string, interface{}, *map[int]RespCode)
}

var (
	RTCPATH = "/usr/local/gortc/"
)

func FullPath(path string) string {
	if path == "" {
		return path
	}

	if strings.HasPrefix(path, "/") {
		return path
	}

	if strings.HasSuffix(RTCPATH, "/") {
		return RTCPATH + path
	} else {
		return RTCPATH + "/" + path
	}
}
