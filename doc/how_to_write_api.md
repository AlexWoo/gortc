# How to write api
---
## Base Example

	package main
	
	import (
		"net/http"
		"rtclib"
	)
	
	type V1 struct {
	}
	
	func APIInstance() rtclib.API {
		return &V1{}
	}
	
	func (i *V1) Get(req *http.Request, paras string) (int,
		*map[string]string, interface{}, *map[int]rtclib.RespCode) {
	
		return 0, nil, "Hello", nil
	}
	
	func (i *V1) Post(req *http.Request, paras string) (int,
		*map[string]string, interface{}, *map[int]rtclib.RespCode) {
	
		return 2, nil, "Hello", nil
	}
	
	func (i *V1) Delete(req *http.Request, paras string) (int,
		*map[string]string, interface{}, *map[int]rtclib.RespCode) {
	
		return 2, nil, nil, nil
	}

## Plugin Interface

Every API must include a plugin interface func named APIInstance, it's protype is

	func APIInstance() rtclib.API {
		return ...
	}

The func return user defined api instance, user defined API must be a struct implementing API interface

## API Interface

	// API interface
	type API interface {
		Get(req *http.Request, paras string) (int,
			*map[string]string, interface{}, *map[int]RespCode)
		Post(req *http.Request, paras string) (int,
			*map[string]string, interface{}, *map[int]RespCode)
		Delete(req *http.Request, paras string) (int,
			*map[string]string, interface{}, *map[int]RespCode)
	}

- Get is entry to deal with GET method
- Post is entry to deal with POST method
- Delete is entry to deal with DELETE method

In these entry:

- req is http request call api
- paras is request uri paras part in the request
- return value1 is user or system defined code. if return value1 is -1, return value3 must be string, the reponse to client is pure string. otherwise, return value3 can be string, map or struct, the response to client is json.
- return value2 is user defined http header to client
- return value3 is user return content to client
- return value4 is user defined response code table,

## Resp Code

System reponse code table

	var syscode = map[int]rtclib.RespCode{
		-1: {Status: 200, Msg: ""},
		0:  {Status: 200, Msg: "OK"},
		1:  {Status: 404, Msg: "Invalid uri format"},
		2:  {Status: 404, Msg: "Unsuppoted Method"},
		3:  {Status: 404, Msg: "Unsuppoted API"},
		4:  {Status: 500, Msg: "API Error"},
		5:  {Status: 500, Msg: "Unsuppoted err code"},
		6:  {Status: 500, Msg: "Unsuppoted ret body"},
	}

If user defined response code table conflict with system response code table, system will use user defined code table in priority

If return code not defined in system response code table or user defined code table, system will return syscode[5] to client

## Return Value
- return value1

	Must in system response code table or user defined response code table which is return value4.
	
	If return value1 is -1, return value3 must be string, the response to user is string. Otherwise, return value3 could be string, map or struct, the response to user is json.

	Apiserver will return indicated status and msg to client in system response code table or user defined response code table.

- return value2

	User defined response headers

- return value3

	if return value1 is not -1, return value3 is:
	
	- string: will replace msg in json using value3
	- map: will append map as json in return json msg
	- struct: will append struct as json in return json msg

- return value4

	user defined response code table, if set nil, use system response code only