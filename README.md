# GO RTC Server
---
## 简介
GO RTC Server 是一个 RTC 信令控制服务器，使用协议为一个类 SIP 协议，为更便于和 web 对接，该协议目前底层传输使用 websocket，协议封装使用 JSON，我们将该协议称为 JSIP 协议。该协议的基本介绍见[JSIP 信令基础规范](https://github.com/AlexWoo/doc/blob/master/直播分发技术/JSIP%20信令基础规范.md)。

GO RTC Server 对外暴露两个服务器，一个是 API 服务器，提供给外部控制和查询使用，一个是 RTC 服务器，处理 JSIP 消息。考虑到 RTC 业务后续的多样性，我们将 GO RTC Server 拆分成了两层：系统层和业务层

- 系统层对外提供 API 服务和 RTC 服务，提供 JSIP 协议栈，以及消息分发，业务逻辑选择
- 业务层是各种 RTC 业务的实现，这些业务是独立于 GO RTC Server 项目的业务插件，这些插件通过编译以后生成 .so。这些 .so 可以通过 Restful API 接口加载到 GO RTC Server 中，从而实现业务逻辑的在线加载和更新

## 开始
### 系统安装

	cd go-rtc-server
	./install --prefix=/usr/local/gortc

如果不指定安装路径，默认安装在 ~/gortc 目录下，完成安装后，在安装目录下会有以下内容：

- bin：可执行文件存放目录

	- gortc：go rtc server 可执行文件
	- pcompile：API 和业务逻辑插件编译脚本

- certs：证书文件存放目录
- conf：配置文件存放目录
- logs：日志文件存放目录
- pkg：依赖库存放目录，编译打包环境不可删除，线上运行环境可以删除
- plugins：API 和业务逻辑插件存放目录
- src：源代码存放路径，编译打包环境不可删除，线上运行环境可以删除

### 系统配置

GO RTC Server 配置文件使用安装目录下的 conf/gortc.ini 作为配置文件，详细内容可见该文件中的详细介绍

### 系统启停

- 系统启动：

	InstallPath/bin/gortc

	默认 APIServer 开启非 https 端口 2539，RTCServer 开启非 https 端口 8080

- 系统优雅停止

	kill -INT pid
	kill -QUIT pid

	pid 为 gortc 进程 ID，使用方式停止，系统会先关闭内部所有正在运行的模块，然后才会整体退出。为了防止系统退出异常而无限期挂起，设置了 5s 定时器，如果指定时间内未完全关闭，系统将强制关闭

- 系统强制关闭

	kill -TERM pid

	pid 为 gortc 进程 ID，系统会直接退出

- 系统配置重加载

	kill -HUP pid

	pid 为 gortc 进程 ID，系统会对可重载的配置进行重载

- 系统日志重打开

	kill -USR1 pid

	pid 为 gortc 进程 ID，系统会关闭已打开的日志文件句柄，并重新打开日志文件，主要提供给日志切分使用

- 系统运行状态查询

	curl http://ip:apiport/runtime/v1/stack

	使用该接口可以看到当前系统堆栈情况，可以通过该命令确定当前是否有非预期的协程挂起，该 Restfule 接口是内置接口，不可删除

### API 加载

API 的更详细介绍相见 [how to write api](doc/how_to_write_api.md)

- 编译

	InstallPath/bin/pcompile path\_to\_apiplugin
	
	这里需要注意，需要指定 API 插件所在目录，这个目录不能以 / 结束，如 ./slpdemo 目录下是当前要编译的 API 插件，编译命令为：
	
	InstallPath/bin/pcompile slpdemo
	
	但不能是
	
	InstallPath/bin/pcompile slpdemo/

- 加载

	curl -XPOST http://ip:apiport/apim/v1/xxx.vn?file=xxx.so

	如上面例子中我们使用
	
	InstallPath/bin/pcompile slpdemo
	
	最终会在 InstallPath/plugins 目录下生成一个 slpdemo.so 文件，我们为这个 API 指定 api 名称为 slpdemo(可以是其它名字)，这个 API 的版本是 v1 版，我们可以使用以下命令进行加载：
	
	curl -XPOST http://ip:apiport/apim/v1/slpdemo.v1?file=slpdemo.so

	完成加载后，我们可以通过：
	
	http://ip:apiport/slpdemo/v1/XXX
	
	访问这个 API，可以通过 POST，DELETE 或 GET 方法访问这个 API 不同的方法接口

- 查看

	curl http://ip:apiport/apim/v1/apis

	可以查看当前系统加载的所有 API 接口

- 删除

	curl -XDELETE http://ip:apiport/apim/v1/apiname

	可以删除指定的 API 接口，如：
	
	curl -XDELETE http://ip:apiport/apim/v1/slpdemo.v1

	注：内置接口不能删除

### 业务逻辑加载

业务逻辑的更详细介绍相见 [how to write slp](doc/how_to_write_slp.md)

- 编译

	编译方法与 API 接口插件编译相同，对于一个插件可以同时是业务逻辑插件和 API 接口插件

- 加载

	curl -XPOST http://ip:apiport/slpm/v1/xxx.vn?file=xxx.so

	如上面例子中我们使用
	
	InstallPath/bin/pcompile slpdemo
	
	最终会在 InstallPath/plugins 目录下生成一个 slpdemo.so 文件，我们为这个业务逻辑指定 a名称为 slpdemo(可以是其它名字)，我们可以使用以下命令进行加载：
	
	curl -XPOST http://ip:apiport/apim/v1/slpdemo?file=slpdemo.so

- 查看

	curl http://ip:apiport/slpm/v1/slps

	可以查看当前系统加载的所有业务逻辑

- 删除

	curl -XDELETE http://ip:apiport/slpm/v1/slpname

	可以删除指定的业务逻辑，如：
	
	curl -XDELETE http://ip:apiport/apim/v1/slpdemo.v1

## 问题

目前 GO 语言 plugin 实现上的一些问题，API，业务逻辑插件在线更新会存在一些问题：在插件重新编译后，插件的文件名不会变化，GO 语言在加载插件时会以文件名为索引，因此不会打开更新后的文件，从而导致无法完成更新。

解决方法：编译后，将插件进行更名，然后使用新的文件名进行加载