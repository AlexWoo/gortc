// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//
// Api Request for apiserver

package rtclib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/tidwall/gjson"
)

// API Request Struct
type APIRequest struct {
	method string
	url    string
	header map[string]string
	body   *map[string]interface{}
	c      *http.Client
}

// Get body from response or request
func APIBody(r io.Reader) (map[string]interface{}, error) {
	body, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	if !gjson.ValidBytes(body) {
		return nil, fmt.Errorf("body must be json, %s", string(body))
	}

	res, ok := gjson.ParseBytes(body).Value().(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("body must be json object, %s", string(body))
	}

	return res, nil
}

// New API Request
func NewAPIRequest(method string, url string, header map[string]string,
	body *map[string]interface{}, timeout time.Duration) *APIRequest {

	return &APIRequest{
		method: method,
		url:    url,
		header: header,
		body:   body,
		c: &http.Client{
			Timeout: timeout,
		},
	}
}

// Send API Request, and Receive API Response
func (r *APIRequest) Do() (*APIResponse, error) {
	var reader io.Reader
	var req *http.Request
	if r.body != nil {
		body, _ := json.Marshal(r.body)
		reader = bytes.NewReader(body)
	}

	// New HTTP Request
	req, err := http.NewRequest(r.method, r.url, reader)
	if err != nil {
		return nil, err
	}

	// Set HTTP Request header
	for k, v := range r.header {
		req.Header.Set(k, v)
	}

	// Send request and wait response
	resp, err := r.c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	apiresp := &APIResponse{
		resp: resp,
	}

	res, err := APIBody(resp.Body)
	if err != nil {
		return nil, err
	}

	code, ok := res["code"].(float64)
	if !ok {
		return apiresp, fmt.Errorf("body has no code, %v", res)
	}

	apiresp.code = int(code)
	apiresp.body = res

	return apiresp, nil
}

// API Response Struct
type APIResponse struct {
	resp *http.Response
	code int
	body map[string]interface{}
}

// get response code
func (r *APIResponse) StatusCode() int {
	return r.resp.StatusCode
}

// get response header
func (r *APIResponse) Header(key string) string {
	return r.resp.Header.Get(key)
}

// get response body
func (r *APIResponse) Body() map[string]interface{} {
	return r.body
}

// get api return code
func (r *APIResponse) APICode() int {
	return r.code
}
