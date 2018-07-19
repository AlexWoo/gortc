// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//
// API Response

package apimodule

import (
	"encoding/json"
	"net/http"
	"reflect"
	"strings"
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
	body    map[string]interface{}
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
	resp := &Response{
		status:  c.Status,
		headers: make(map[string]string),
		body:    make(map[string]interface{}),
	}

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
	resp.body["code"] = code
	resp.body["msg"] = msg

	if body == nil {
		return
	}

	typ := reflect.ValueOf(body).Kind()

	if typ == reflect.String {
		resp.body["msg"] = body
		return
	}

	if typ == reflect.Map {
		for _, k := range reflect.ValueOf(body).MapKeys() {
			key := strings.ToLower(k.String())
			val := reflect.ValueOf(body).MapIndex(k).Interface()
			resp.body[key] = val
		}
		return
	}

	if typ == reflect.Struct {
		t := reflect.TypeOf(body)
		v := reflect.ValueOf(body)
		for i := 0; i < t.NumField(); i++ {
			key := strings.ToLower(t.Field(i).Name)
			val := v.Field(i).Interface()
			resp.body[key] = val
		}

		return
	}

	//TODO logErr
	resp.status = syscode[5].Status
	resp.body["code"] = 5
	resp.body["msg"] = syscode[5].Msg
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
	code, _ := resp.body["code"].(int)
	if code == -1 {
		body, _ := resp.body["msg"].(string)
		w.Write([]byte(body))
	} else {
		body, _ := json.Marshal(resp.body)
		w.Write(body)
	}
}
