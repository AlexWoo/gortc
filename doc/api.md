## 1.1 系统运行状态

本接口用于查看系统运行时状态

*接口:* ***/runtime/v1/stack***

***请求URL参数说明:***

无

***请求头参数说明:***

无

***请求方法:***

GET

***请求体参数说明:***

无

***响应参数说明***

无

***参考请求:***

	curl http://127.0.0.1:2539/runtime/v1/stack

***参考响应:***

	goroutine 19 [running]:
	main.stack(0x890e78, 0xd23120)
		/usr/local/gortc/src/gortc/runtime.v1.go:25 +0xbb
	main.(*RUNTIME_V1).Get(0xd6f918, 0xc420360100, 0xc4203441d0, 0x5, 0xd6f918, 0xc4203441cd, 0x2, 0xc42004db40, 0xa)
		/usr/local/gortc/src/gortc/runtime.v1.go:41 +0x82
	main.(*apiServer).callAPI(0xc42004e240, 0xc420360100, 0xc4203441c5, 0x7, 0xc4203441cd, 0x2, 0xc4203441d0, 0x5, 0xc4203441d0, 0x5, ...)
		/usr/local/gortc/src/gortc/apiserver.go:116 +0x262
	main.(*apiServer).handler(0xc42004e240, 0xd230a0, 0xc4200ff1e0, 0xc420360100)
		/usr/local/gortc/src/gortc/apiserver.go:133 +0xb1
	main.(*apiServer).(main.handler)-fm(0xd230a0, 0xc4200ff1e0, 0xc420360100)
		/usr/local/gortc/src/gortc/apiserver.go:160 +0x48
	github.com/alexwoo/golib.(*HTTPServer).handler(0xc42006c870, 0xd234e0, 0xc420362000, 0xc420360100)
		/usr/local/gortc/src/github.com/alexwoo/golib/httpserver.go:199 +0x236
	github.com/alexwoo/golib.(*HTTPServer).(github.com/alexwoo/golib.handler)-fm(0xd234e0, 0xc420362000, 0xc420360100)
		/usr/local/gortc/src/github.com/alexwoo/golib/httpserver.go:157 +0x48
	net/http.HandlerFunc.ServeHTTP(0xc4200112b0, 0xd234e0, 0xc420362000, 0xc420360100)
		/usr/lib/golang/src/net/http/server.go:1918 +0x44
	net/http.(*ServeMux).ServeHTTP(0xc420061e60, 0xd234e0, 0xc420362000, 0xc420360100)
		/usr/lib/golang/src/net/http/server.go:2254 +0x130
	net/http.serverHandler.ServeHTTP(0xc4200fea90, 0xd234e0, 0xc420362000, 0xc420360100)
		/usr/lib/golang/src/net/http/server.go:2619 +0xb4
	net/http.(*conn).serve(0xc42010a0a0, 0xd23b60, 0xc42004e7c0)
		/usr/lib/golang/src/net/http/server.go:1801 +0x71d
	created by net/http.(*Server).Serve
		/usr/lib/golang/src/net/http/server.go:2720 +0x288
	
	goroutine 1 [select]:
	github.com/alexwoo/golib.(*Modules).mainloop(0xc42004e200)
		/usr/local/gortc/src/github.com/alexwoo/golib/module.go:195 +0x1a0
	github.com/alexwoo/golib.(*Modules).Start(0xc42004e200)
		/usr/local/gortc/src/github.com/alexwoo/golib/module.go:101 +0x1b9
	main.main()
		/usr/local/gortc/src/gortc/gortc.go:34 +0x7cc
	……

## 1.2 API 管理

### 1.2.1 API 查询

本接口用于查询系统中所有的 API

*接口:* ***/apim/v1/apis***

***请求URL参数说明:***

无

***请求头参数说明:***

无

***请求方法:***

GET

***请求体参数说明:***

无

***响应参数说明***

无

***参考请求:***

	curl http://127.0.0.1:2539/apim/v1/apis

***参考响应:***

	api		file
	------------------------------------------------------------
	chatroom.v1	chatroom.so
	serviceregistry.v1	srvreg.so
	runtime.v1
	apim.v1
	slpm.v1
	------------------------------------------------------------

没有对应插件名的 api 为系统内置 api，不能删除

### 1.2.2 API 加载

本接口用于向系统中添加 API

*接口:* ***/apim/v1/\<apiname.version\>?file=\<apiplugin\>***

***请求URL参数说明:***

无

***请求头参数说明:***

无

***请求方法:***

POST

***请求体参数说明:***

无

***响应参数说明***

无

***参考请求:***

	curl -XPOST http://127.0.0.1:2539/apim/v1/chatroom.v1?file=chatroom.so

chatroom.so 为使用 pcompile 编译后，在 /path/to/gortc/plugins 目录下生成的插件

***参考响应:***

	Load API chatroom.v1 chatroom.so successd

加载成功，系统中将会增加名为 chatroom， version 为 v1 的 API，可以使用 /chatroom/v1 调用该 API

### 1.2.3 API 删除

本接口用于删除系统中的 API

*接口:* ***/apim/v1/\<apiname.version\>***

***请求URL参数说明:***

无

***请求头参数说明:***

无

***请求方法:***

DELETE

***请求体参数说明:***

无

***响应参数说明***

无

***参考请求:***

	curl -XDELETE http://127.0.0.1:2539/apim/v1/chatroom.v1

***参考响应:***

	Delete API chatroom.v1 successd

## 1.3 业务逻辑管理

### 1.3.1 业务逻辑查询

本接口用于查询系统中的所有业务逻辑

*接口:* ***/slpm/v1/slps***

***请求URL参数说明:***

无

***请求头参数说明:***

无

***请求方法:***

GET

***请求体参数说明:***

无

***响应参数说明***

无

***参考请求:***

	curl http://127.0.0.1:2539/slpm/v1/slps

***参考响应:***

	slp		used		using		file		time
	------------------------------------------------------------
	default	0	1	chatroom.so	2018-10-10 16:11:55.062
	------------------------------------------------------------

### 1.3.2 API 加载

本接口向系统中添加新的业务逻辑

*接口:* ***/slpm/v1/\<slpname\>?file=\<slpplugin\>***

***请求URL参数说明:***

无

***请求头参数说明:***

无

***请求方法:***

POST

***请求体参数说明:***

无

***响应参数说明***

无

***参考请求:***

	curl -XPOST http://127.0.0.1:2539/slpm/v1/chatroom?file=chatroom.so

chatroom.so 为使用 pcompile 编译后，在 /path/to/gortc/plugins 目录下生成的插件

***参考响应:***

	Load SLP chatroom chatroom.so successd

### 1.3.3 API 删除

本接口用于删除系统中的业务逻辑

*接口:* ***/slpm/v1/\<slpname\>***

***请求URL参数说明:***

无

***请求头参数说明:***

无

***请求方法:***

DELETE

***请求体参数说明:***

无

***响应参数说明***

无

***参考请求:***

	curl -XDELETE http://127.0.0.1:2539/slpm/v1/chatroom

***参考响应:***

	Delete SLP chatroom successd