# How to write slp
---

## Basic elems for writing SLP

### Plugin Interface

Every SLP must include a plugin interface func named GetInstance, it's protype is

	func GetInstance(task *rtclib.Task) rtclib.SLP {
		return ...
	}

The func return user's slp instance, user's SLP must be a struct implementing SLP interface

### SLP Interface

	type SLP interface {
		// Task start in normal stage, msg process interface
		Process(jsip *JSIP)
	
		// Task start in SLP loaded in gortc, msg process interface
		OnLoad(jsip *JSIP)
	
		// Create SLP ctx
		NewSLPCtx() interface{}
	}

- Process

	An incoming session will create a normal slp instance. msg send to the slp instance will entry from Process

- OnLoad

	A onload slp instance will create when slp load in go rtc server, msg send to the slp instance will entry from OnLoad

- NewSLPCtx

	NewSLPCtx will create a ctx shared in all instances of current SLP

### SLP directory

	userslp
		|--- slp codes
		|--- conf
				|--- slp confs

After compile slp

	gortc
		|--- plugins
		|		|--- userslp.so
		|--- conf
				|--- userslp
						|--- slp confs

A .so file will generated under gortc installpath/plugins directory. The .so file name is basename of userslp's directory.

A conf diretory will gernerated under gortc installpath/conf diretory. The conf diretory is basename of userslp's directory. If same conf file is under install conf path, the file under slp conf diretory will not cover the file under install conf path

Examples:

gortc server install under /usr/local/gortc, userslp path is ./slpdemo, it has a conf path ./slpdemo/conf. When will execute pcompile ./slpdemo, it will generate a slpdemo.so under /usr/local/gortc/plugins directory, and a slpdemo diretory under /usr/local/gortc/conf/ directory.

## SLP Programing Interface

SLP Programing interface defined in rtclib, user should import "rtclib" in user's code

### JSIP interface

Defined in rtclib/jsip.go

#### JSIP

**STRUCT**

	type JSIP struct {
		Type       int
		Code       int
		RequestURI string
		From       string
		To         string
		CSeq       uint64
		DialogueID string
		Router     []string
		Body       string
		RawMsg     map[string]interface{}
	
		inner       bool
		conn        golib.Conn
		Transaction *JSIPTrasaction
		Session     *JSIPSession
	}

- Type is JSIP msg type, it's value will be:

	- INVITE
	- ACK
	- BYE
	- CANCEL
	- REGISTER
	- OPTIONS
	- INFO
	- UPDATE
	- PRACK
	- SUBSCRIBE
	- MESSAGE
	- TERM

	TERM is internal msg to indicator session is terminate, other msg is same as SIP

- Code

	0 present the msg is a jsip request, other code present msg is a jsip response, other code is same as SIP

- RequestURI, From, To

	Same as SIP

- CSeq

	An integer to present differnt transactions in session

- DialogueID

	An string to present differnt session in system, it's same as Call-ID + From-tag + To-tag in SIP

- Body

	User Content

- RawMsg

	Raw JSIP Msg receive from websocket channel

- Transaction

	JSIP transaction

- Session

	JSIP session

**FUNCTIONS**

	func (jsip *JSIP) Name() string

JSIP msg name, such as INVITE, INVITE_200

	func (jsip *JSIP) Abstract() string

Include JSIP msg name and some important header

	func (jsip *JSIP) Detail() string

Include JSIP msg name and all content in jsip msg

	func (jsip *JSIP) GetInt(header string) (int64, bool)

Get a integer type header

	func (jsip *JSIP) SetInt(header string, value int64)

Set a integer type header

	func (jsip *JSIP) GetString(header string) (string, bool)

Get a string type header

	func (jsip *JSIP) SetString(header string, value string)

Set a string type header

	func SendMsg(jsip *JSIP)

Send a jsip msg

	func SendJSIPTerm(dlg string)

Send a jsip term msg to terminate jsip session, it will translate to correct msg according to session state

	func JSIPMsgClone(req *JSIP, dlg string) *JSIP

Clone a jsip msg

	func JSIPMsgRes(req *JSIP, code int) *JSIP

Generate a response for a request msg

	func JSIPMsgAck(resp *JSIP) *JSIP

Generate a ack for a response msg

	func JSIPMsgBye(session *JSIPSession) *JSIP

Generate a bye for a jsip session

	func JSIPMsgUpdate(session *JSIPSession) *JSIP

Generate a update for a jsip session

	func JSIPMsgCancel(req *JSIP) *JSIP

Generate a cancel for a jsip session

#### Other

**STRUCT**

	type JSIPUri struct {
		UserWithHost   string
		UserWithPrefix string
		HostWithPort   string
		Prefix         string
		User           string
		Host           string
		Port           uint16
		Paras          map[string]interface{} // string or bool
	}

JSIP Uri format is

	[[prefix:]user@]host[:port][;para1=value][;para2]

- UserWithHost

	[[prefix:]user@]host

- UserWithPrefix

	[[prefix:]user

- HostWithPort

	host[:port]

**FUNCTIONS**

	func ParseJSIPUri(uri string) (*JSIPUri, error)

Parse a jsip uri to JSIPUri struct

### task interface

Defined in rtclib/task.go

	func (t *Task) NewDialogueID() string 

New a jsip dialogue, using for slp sending a new jsip session. This func will generate a new dialogue and related the dialogue to slp, the new jsip session msg can be send to the slp instance then.

	func (t *Task) GetCtx() interface{} 

Return ctx shared in all instances of same SLP.

	func (t *Task) SetFinished()

Tell go rtc server, current slp instance is finish, go rtc server will recycle the slp instance then.

	func (t *Task) LogDebug(format string, v ...interface{})

log a debug level log

	func (t *Task) LogInfo(format string, v ...interface{})

log a info level log

	func (t *Task) LogError(format string, v ...interface{})

log a error level log

	func (t *Task) LogFatal(format string, v ...interface{})

log a fatal level log, it will cause go rtc server exit

### other interface

Defined in rtclib/base.go

	func FullPath(path string) string 

complete path with GORTC install path, example:

go rtc server install under /usr/local/gortc

FullPath("conf/slpdemo") will get path /usr/local/gortc/conf/slpdemo

It provides to slp to get conf path

## JSIP process in go rtc server

How JSIP Msg process in go rtc server, introduced in [slp process in go rtc server](slp_process_in_go-rtc-server.md)
