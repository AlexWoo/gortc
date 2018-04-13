// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//
// API Response

package apimodule

import (
	"net/http"
	"reflect"
	"strings"

	simplejson "github.com/bitly/go-simplejson"
)

var syscode = map[int]RespCode{
	-1: {Status: 200, Msg: ""},
	0:  {Status: 200, Msg: "OK"},
	1:  {Status: 404, Msg: "Invalid uri format"},
	2:  {Status: 404, Msg: "Unsuppoted Method"},
	3:  {Status: 404, Msg: "Unsuppoted API"},
	4:  {Status: 500, Msg: "API Error"},
	5:  {Status: 500, Msg: "Unsuppoted err code"},
	6:  {Status: 500, Msg: "Unsuppoted ret body"},
}

type RespCode struct {
	Status int
	Msg    string
}

type Response struct {
	status  int
	headers map[string]string
	body    *simplejson.Json
}

func NewResponse(code int, headers *map[string]string, body interface{},
	usercode *map[int]RespCode) *Response {

	// respcode
	var c RespCode
	ok := false
	if usercode != nil { /* usercode set */
		c, ok = (*usercode)[code]
	}
	if !ok { /* usercode not set */
		c, ok = syscode[code]
	}
	if !ok {
		return NewResponse(4, nil, nil, nil)
	}

	// init resp
	resp := &Response{status: c.Status}
	resp.headers = make(map[string]string)
	resp.headers["Server"] = "RTC-APIServer"
	resp.headers["Content-Type"] = "application/json; charset=utf-8"

	// userheaders
	if headers != nil {
		for k, v := range *headers {
			resp.headers[k] = v
		}
	}

	// body
	resp.setBody(code, c.Msg, body)

	return resp
}

func (resp *Response) setBody(code int, msg string, body interface{}) {
	resp.body = simplejson.New()
	resp.body.Set("code", code)
	resp.body.Set("msg", msg)

	if body == nil {
		return
	}

	typ := reflect.ValueOf(body).Kind()

	if typ == reflect.String {
		resp.body.Set("msg", body)
		return
	}

	if typ == reflect.Map {
		for _, k := range reflect.ValueOf(body).MapKeys() {
			key := strings.ToLower(k.String())
			val := reflect.ValueOf(body).MapIndex(k).Interface()
			resp.body.Set(key, val)
		}
		return
	}

	if typ == reflect.Struct {
		t := reflect.TypeOf(body)
		v := reflect.ValueOf(body)
		for i := 0; i < t.NumField(); i++ {
			key := strings.ToLower(t.Field(i).Name)
			val := v.Field(i).Interface()
			resp.body.Set(key, val)
		}

		return
	}

	//TODO logErr
	resp.status = syscode[5].Status
	resp.body.Set("code", 5)
	resp.body.Set("msg", syscode[5].Msg)
	return
}

func (resp *Response) SendResp(w http.ResponseWriter) {
	// resp headers
	for h, v := range resp.headers {
		w.Header().Set(h, v)
	}

	// resp status
	w.WriteHeader(resp.status)

	// send body
	code, _ := resp.body.Get("code").Int()
	if code == -1 {
		body, _ := resp.body.Get("msg").Bytes()
		w.Write(body)
	} else {
		body, _ := resp.body.MarshalJSON()
		w.Write(body)
	}
}
