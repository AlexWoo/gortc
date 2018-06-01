// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//
// RTC Json SIP

package rtclib

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"

	simplejson "github.com/bitly/go-simplejson"
)

// return value
const (
	ERROR = iota
	OK
	IGNORE
)

// SIP Request
const (
	UNKNOWN = iota
	INVITE
	ACK
	BYE
	CANCEL
	REGISTER
	OPTIONS
	INFO
	UPDATE
	PRACK
	SUBSCRIBE
	MESSAGE
)

var jsipReqUnparse = map[string]int{
	"INVITE":    INVITE,
	"ACK":       ACK,
	"BYE":       BYE,
	"CANCEL":    CANCEL,
	"REGISTER":  REGISTER,
	"OPTIONS":   OPTIONS,
	"INFO":      INFO,
	"UPDATE":    UPDATE,
	"PRACK":     PRACK,
	"SUBSCRIBE": SUBSCRIBE,
	"MESSAGE":   MESSAGE,
}

var jsipReqParse = map[int]string{
	INVITE:    "INVITE",
	ACK:       "ACK",
	BYE:       "BYE",
	CANCEL:    "CANCEL",
	REGISTER:  "REGISTER",
	OPTIONS:   "OPTIONS",
	INFO:      "INFO",
	UPDATE:    "UPDATE",
	PRACK:     "PRACK",
	SUBSCRIBE: "SUBSCRIBE",
	MESSAGE:   "MESSAGE",
}

var jsipResDesc = map[int]string{
	100: "Trying",
	180: "Ringing",
	181: "Call Is Being Forwarded",
	182: "Queued",
	183: "Session Progress",
	200: "OK",
	202: "Accepted",
	300: "Multiple Choices",
	301: "Moved Permanently",
	302: "Moved Temporarily",
	305: "Use Proxy",
	380: "Alternative Service",
	400: "Bad Request",
	401: "Unauthorized",
	402: "Payment Required",
	403: "Forbidden",
	404: "Not Found",
	405: "Method Not Allowed",
	406: "Not Acceptable",
	407: "Proxy Authentication Required",
	408: "Request Timeout",
	410: "Gone",
	413: "Request Entity Too Large",
	414: "Request-URI Too Large",
	415: "Unsupported Media Type",
	416: "Unsupported URI Scheme",
	420: "Bad Extension",
	421: "Extension Required",
	423: "Interval Too Brief",
	480: "Temporarily not available",
	481: "Call Leg/Transaction Does Not Exist",
	482: "Loop Detected",
	483: "Too Many Hops",
	484: "Address Incomplete",
	485: "Ambiguous",
	486: "Busy Here",
	487: "Request Terminated",
	488: "Not Acceptable Here",
	491: "Request Pending",
	493: "Undecipherable",
	500: "Internal Server Error",
	501: "Not Implemented",
	502: "Bad Gateway",
	503: "Service Unavailable",
	504: "Server Time-out",
	505: "SIP Version not supported",
	513: "Message Too Large",
	600: "Busy Everywhere",
	603: "Decline",
	604: "Does not exist anywhere",
	606: "Not Acceptable",
}

// SIP Transaction
const (
	TRANS_REQ = iota
	TRANS_TRYING
	TRANS_PR
	TRANS_FINALRESP
)

// SIP Session
const (
	INVITE_INIT = iota
	INVITE_REQ
	INVITE_18X
	INVITE_PRACK
	INVITE_UPDATE
	INVITE_200
	INVITE_ACK
	INVITE_REINV
	INVITE_RE200
	INVITE_ERR
	INVITE_END
)

var jsipInviteState = map[int]string{
	INVITE_INIT:   "INVITE_INIT",
	INVITE_REQ:    "INVITE_REQ",
	INVITE_18X:    "INVITE_18X",
	INVITE_PRACK:  "INVITE_PRACK",
	INVITE_UPDATE: "INVITE_UPDATE",
	INVITE_200:    "INVITE_200",
	INVITE_ACK:    "INVITE_ACK",
	INVITE_REINV:  "INVITE_REINV",
	INVITE_RE200:  "INVITE_RE200",
	INVITE_ERR:    "INVITE_ERR",
	INVITE_END:    "INVITE_END",
}

const (
	DEFAULT_INIT = iota
	DEFAULT_REQ
	DEFAULT_RESP
)

const (
	UAS = iota
	UAC
)

const (
	RECV = iota
	SEND
)

var jsipDirect = map[int]string{
	RECV: "RECV",
	SEND: "SEND",
}

type JSIP struct {
	Type       int
	Code       int
	RequestURI string
	From       string
	To         string
	CSeq       uint64
	DialogueID string
	Router     []string
	Body       interface{}
	RawMsg     map[string]interface{}
}

type JSIPTrasaction struct {
	typ    int
	state  int
	uatype int
	req    *JSIP
	cseq   uint64
}

type JSIPSession struct {
	conn    *JSIPConn
	typ     int
	state   int
	uatype  int
	req     *JSIP
	cseq    uint64
	handler func(session *JSIPSession, jsip *JSIP, sendrecv int) int
}

var (
	Jtransactions = make(map[string]*JSIPTrasaction)
	Jsessions     = make(map[string]*JSIPSession)
	JsessLock     sync.RWMutex
	JtransLock    sync.RWMutex
)

var (
	jsipHandle  func(jsip *JSIP)
	rtclog      *Log
	realm       string
	rtclocation string
)

func JsessGet(dlg string) *JSIPSession {
	JsessLock.RLock()
	defer JsessLock.RUnlock()
	return Jsessions[dlg]
}

func JsessSet(dlg string, session *JSIPSession) {
	JsessLock.Lock()
	defer JsessLock.Unlock()
	Jsessions[dlg] = session
}

func JsessDel(dlg string) {
	JsessLock.Lock()
	defer JsessLock.Unlock()
	delete(Jsessions, dlg)
}

func JtransGet(tid string) *JSIPTrasaction {
	JtransLock.RLock()
	defer JtransLock.RUnlock()
	return Jtransactions[tid]
}

func JtransSet(tid string, trans *JSIPTrasaction) {
	JtransLock.Lock()
	defer JtransLock.Unlock()
	Jtransactions[tid] = trans
}

func JtransDel(tid string) {
	JtransLock.Lock()
	defer JtransLock.Unlock()
	delete(Jtransactions, tid)
}

func InitHandler(h func(jsip *JSIP), log *Log, rlm string, location string) {
	jsipHandle = h
	rtclog = log
	realm = rlm
	rtclocation = location
}

func transactionID(jsip *JSIP, cseq uint64) string {
	return jsip.DialogueID + "_" + strconv.FormatUint(cseq, 10)
}

// interface
func JsipParseUri(uri string) (string, string, []string) {
	var user, host string
	var paras []string

	ss := strings.Split(uri, "@")
	if len(ss) == 1 {
		host = ss[0]
	} else if len(ss) == 2 {
		user = ss[0]
		host = ss[1]
	} else {
		return user, host, paras
	}

	ss = strings.Split(host, ";")
	if len(ss) > 1 {
		host = ss[0]
		paras = ss[1:]
	}

	return user, host, paras
}

func JsipName(jsip *JSIP) string {
	req := jsipReqParse[jsip.Type]
	if req == "" {
		req = "UNKNOWN"
	}

	if jsip.Code == 0 {
		return req
	} else {
		code := strconv.Itoa(jsip.Code)
		return req + "_" + code
	}
}

func SendJSIPReq(req *JSIP, dlg string) {
	var conn *JSIPConn

	if JsessGet(dlg) == nil {
		if len(req.Router) > 0 {
			// Use Router Header 0 as default route if exist
			conn = jstack.RTCClient(req.Router[0])
		} else {
			// Use Request-URI as default route if Router Header not exist
			conn = jstack.RTCClient(req.RequestURI)
		}

		if conn == nil {
			return
		}
	} else {
		conn = JsessGet(dlg).conn
	}

	req.DialogueID = dlg

	SendJsonSIPMsg(conn, req)
}

func SendJSIPRes(req *JSIP, code int) {
	if req.Code != 0 {
		jstack.log.LogError("Cannot send response for response")
		return
	}

	resp := &JSIP{
		Type:       req.Type,
		Code:       code,
		From:       req.From,
		To:         req.To,
		CSeq:       req.CSeq,
		DialogueID: req.DialogueID,
		RawMsg:     make(map[string]interface{}),
	}

	SendJsonSIPMsg(nil, resp)
}

func SendJSIPAck(resp *JSIP) {
	ack := &JSIP{
		Type:       ACK,
		From:       resp.From,
		To:         resp.To,
		DialogueID: resp.DialogueID,
		RawMsg:     make(map[string]interface{}),
	}

	resp.RawMsg["RelatedID"] = resp.CSeq

	SendJsonSIPMsg(nil, ack)
}

func SendJSIPBye(session *JSIPSession) {
	resp := &JSIP{
		Type:       BYE,
		DialogueID: session.req.DialogueID,
	}

	SendJsonSIPMsg(nil, resp)
}

func SendJSIPCancel(session *JSIPSession, req *JSIP) {
	if req.Type >= 100 {
		jstack.log.LogError("Cannot cancel response")
		return
	}

	resp := &JSIP{
		Type:       CANCEL,
		RequestURI: req.RequestURI,
		From:       req.From,
		To:         req.To,
		DialogueID: session.req.DialogueID,
		RawMsg:     make(map[string]interface{}),
	}

	resp.RawMsg["RelatedID"] = req.CSeq

	SendJsonSIPMsg(nil, resp)
}

// Syntax Layer
func jsipPrepared(jsip *JSIP) (*JSIP, error) {
	if jsip.DialogueID == "" {
		return nil, errors.New("DialogueID not set")
	}

	if jsipReqParse[jsip.Type] == "" {
		return nil, errors.New("Unknown message type")
	}

	if jsip.Code != 0 && (jsip.Code < 100 || jsip.Code > 699) {
		return nil, fmt.Errorf("Unknown response %s", JsipName(jsip))
	}

	if jsip.RawMsg == nil {
		jsip.RawMsg = make(map[string]interface{})
	}

	session := JsessGet(jsip.DialogueID)
	if session == nil {
		if jsip.Code != 0 {
			return nil, errors.New("Cannot send Response for a new session")
		}

		if jsip.RequestURI == "" || jsip.From == "" || jsip.To == "" {
			return nil, errors.New("Must header not set for first request")
		}

		jsip.CSeq = 1
	} else {
		if jsip.RequestURI == "" {
			jsip.RequestURI = session.req.RequestURI
		}

		if jsip.From == "" {
			jsip.From = session.req.From
		}

		if jsip.To == "" {
			jsip.To = session.req.To
		}

		if jsip.Code == 0 { // Request
			session.cseq++
			jsip.CSeq = session.cseq
		} else {
			if jsip.CSeq == 0 {
				return nil, errors.New("Prepared response but CSeq not set")
			}
		}
	}

	if jsip.Code != 0 {
		jsip.RawMsg["Type"] = "RESPONSE"
		jsip.RawMsg["Code"] = jsip.Code
		jsip.RawMsg["Desc"] = jsipResDesc[jsip.Code]
	} else {
		jsip.RawMsg["Type"] = jsipReqParse[jsip.Type]
		jsip.RawMsg["Request-URI"] = jsip.RequestURI
	}

	jsip.RawMsg["From"] = jsip.From
	jsip.RawMsg["To"] = jsip.To
	jsip.RawMsg["DialogueID"] = jsip.DialogueID
	jsip.RawMsg["CSeq"] = jsip.CSeq

	if len(jsip.Router) > 0 {
		router := jsip.Router[0]
		for i := 1; i < len(jsip.Router); i++ {
			router += ", " + jsip.Router[i]
		}
		jsip.RawMsg["Router"] = router
	}

	if jsip.Body != nil {
		jsip.RawMsg["Body"] = jsip.Body
	}

	return jsip, nil
}

func jsipUnParser(data []byte) (*JSIP, error) {
	json, err := simplejson.NewJson(data)
	if err != nil {
		return nil, err
	}

	jsip := &JSIP{}

	typ, err := json.Get("Type").String()
	if err != nil {
		return nil, errors.New("no Type in jsip msg")
	}

	if typ == "RESPONSE" {
		jsip.Code, err = json.Get("Code").Int()
		if err != nil {
			return nil, errors.New("no Code in jsip response")
		}

		if jsip.Code < 100 || jsip.Code > 699 {
			return nil, fmt.Errorf("unexpected status code %d", jsip.Code)
		}
	} else {
		jsip.Type = jsipReqUnparse[typ]
		if jsip.Type == UNKNOWN {
			return nil, errors.New("unexpected Type")
		}

		jsip.RequestURI, err = json.Get("Request-URI").String()
		if err != nil {
			return nil, errors.New("no Request-URI in jsip request")
		}
	}

	jsip.From, err = json.Get("From").String()
	if err != nil {
		return nil, errors.New("no From in jsip message")
	}

	jsip.To, err = json.Get("To").String()
	if err != nil {
		return nil, errors.New("no To in jsip message")
	}

	jsip.DialogueID, err = json.Get("DialogueID").String()
	if err != nil {
		return nil, errors.New("no DialogueID in jsip message")
	}

	jsip.CSeq, err = json.Get("CSeq").Uint64()
	if err != nil {
		return nil, errors.New("no CSeq in jsip message")
	}

	routers, _ := json.Get("Router").String()
	if routers != "" {
		jsip.Router = strings.Split(routers, ",")
		for i := 0; i < len(jsip.Router); i++ {
			jsip.Router[i] = strings.TrimSpace(jsip.Router[i])
		}
	}

	jsip.RawMsg, err = json.Map()

	jsip.Body = json.Get("Body")

	session := JsessGet(jsip.DialogueID)
	if session != nil {
		if jsip.Code == 0 && jsip.CSeq > session.cseq {
			session.cseq = jsip.CSeq
		}
	}

	return jsip, nil
}

// Transaction Layer
func jsipTrasaction(jsip *JSIP, sendrecv int) int {
	tid := transactionID(jsip, jsip.CSeq)
	trans := JtransGet(tid)

	if trans == nil { // Request
		if jsip.Code != 0 {
			jstack.log.LogError("process %s but trans is nil", JsipName(jsip))
			return ERROR
		}

		trans = &JSIPTrasaction{
			typ:   jsip.Type,
			state: TRANS_REQ,
			req:   jsip,
			cseq:  jsip.CSeq,
		}

		JtransSet(tid, trans)

		if sendrecv == RECV {
			trans.uatype = UAS
		} else {
			trans.uatype = UAC
		}

		if jsip.Type == ACK {
			JtransDel(tid)

			relatedid, ok := jsip.RawMsg["RelatedID"]
			if !ok {
				jstack.log.LogInfo("ACK miss RelatedID")
				return IGNORE
			}

			rid, _ := strconv.ParseUint(string(relatedid.(json.Number)), 10, 64)
			tid = transactionID(jsip, rid)
			ackTrans := JtransGet(tid)
			if ackTrans == nil {
				jstack.log.LogInfo("Transaction INVITE not exist")
				return IGNORE
			}

			JtransDel(tid)
		}

		if jsip.Type == CANCEL {
			relatedid, ok := jsip.RawMsg["RelatedID"]
			if !ok {
				jstack.log.LogInfo("CANCEL miss RelatedID")
				return IGNORE
			}

			rid, _ := strconv.ParseUint(string(relatedid.(json.Number)), 10, 64)
			tid = transactionID(jsip, rid)
			cancelTrans := JtransGet(tid)
			if cancelTrans == nil {
				jstack.log.LogInfo("Transaction Cancelled not exist")
				return IGNORE
			}

			if cancelTrans.state == TRANS_FINALRESP {
				jstack.log.LogInfo("Transaction in finalize response, cannot cancel")
				return IGNORE
			}
		}

		return OK
	}

	if jsip.Code == 0 {
		jstack.log.LogError("process %s but trans exist", JsipName(jsip))
		return ERROR
	}

	// Response
	if trans.uatype == UAS && sendrecv == RECV ||
		trans.uatype == UAC && sendrecv == SEND {

		jstack.log.LogError("Response direct is same as Request direct")
		return ERROR
	}

	if jsip.Code == 100 {
		if trans.state > TRANS_TRYING {
			jstack.log.LogError("process 100 Trying but state is %d", trans.state)
			return ERROR
		}

		trans.state = TRANS_TRYING

		return IGNORE
	}

	jsip.Type = trans.typ

	if jsip.Code < 200 && jsip.Code > 100 {
		if trans.state > TRANS_PR {
			jstack.log.LogError("process %s but state is %d", JsipName(jsip),
				trans.state)
			return ERROR
		}

		trans.state = TRANS_PR

		return OK
	}

	if trans.state == TRANS_FINALRESP {
		jstack.log.LogError("process %s but state is %d", JsipName(jsip),
			trans.state)
		return ERROR
	}

	trans.state = TRANS_FINALRESP

	if trans.typ != INVITE {
		JtransDel(tid)
	}

	if trans.typ == CANCEL && sendrecv == RECV {
		// Ignore CANCEL 200 received
		return IGNORE
	}

	if trans.typ == BYE && sendrecv == RECV {
		// Ignore BYE 200 received
		return IGNORE
	}

	return OK
}

// Session Layer
func jsipInviteSession(session *JSIPSession, jsip *JSIP, sendrecv int) int {
	if jsip.Type == INFO {
		return OK
	}
	switch session.state {
	case INVITE_INIT:
		if jsip.Type == INVITE && jsip.Code == 0 {
			session.state = INVITE_REQ
			return OK
		}
	case INVITE_REQ:
		switch jsip.Type {
		case CANCEL:
			return OK
		case INVITE:
			switch {
			case jsip.Code < 200 && jsip.Code > 100:
				session.state = INVITE_18X
				return OK
			case jsip.Code == 200:
				session.state = INVITE_200
				return OK
			case jsip.Code >= 300:
				session.state = INVITE_ERR
				if sendrecv == RECV {
					SendJSIPAck(jsip)
				}
				return OK
			}
		}
	case INVITE_18X:
		switch jsip.Type {
		case CANCEL:
			return OK
		case INVITE:
			switch {
			case jsip.Code < 200 && jsip.Code > 100:
				return OK
			case jsip.Code == 200:
				session.state = INVITE_200
				return OK
			case jsip.Code >= 300:
				session.state = INVITE_ERR
				if sendrecv == RECV {
					SendJSIPAck(jsip)
				}
				return OK
			}
		case PRACK:
			if jsip.Code == 0 && sendrecv == session.uatype {
				session.state = INVITE_PRACK
				return OK
			}
		case UPDATE:
			if jsip.Code == 0 {
				session.state = INVITE_UPDATE
				return OK
			}
		}
	case INVITE_PRACK:
		if jsip.Code == 200 && jsip.Type == PRACK {
			session.state = INVITE_18X
			return OK
		}
	case INVITE_UPDATE:
		if jsip.Code == 200 && jsip.Type == UPDATE {
			session.state = INVITE_18X
			return OK
		}
	case INVITE_200:
		if jsip.Type == ACK {
			session.state = INVITE_ACK
			return OK
		} else if jsip.Type == BYE {
			if sendrecv == RECV {
				SendJSIPRes(jsip, 200)
			}
			session.state = INVITE_END
			return OK
		}
	case INVITE_ACK:
		switch {
		case jsip.Type == INVITE:
			if jsip.Code == 0 {
				session.state = INVITE_REINV
				return OK
			}
		case jsip.Type == UPDATE:
			if jsip.Code == 0 {
				SendJSIPRes(jsip, 200)
				return OK
			}

			if jsip.Code == 200 {
				if sendrecv == SEND {
					return OK
				}
			}
		case jsip.Type == BYE:
			if sendrecv == RECV {
				SendJSIPRes(jsip, 200)
			}
			session.state = INVITE_END
			return OK
		case jsip.Type == INFO: // INFO and INFO 200
			return OK
		}
	case INVITE_REINV:
		if jsip.Code == 200 && jsip.Type == INVITE {
			session.state = INVITE_RE200
			return OK
		} else if jsip.Type == BYE {
			if sendrecv == RECV {
				SendJSIPRes(jsip, 200)
			}
			session.state = INVITE_END
			return OK
		}
	case INVITE_RE200:
		if jsip.Type == ACK {
			session.state = INVITE_ACK
			return OK
		} else if jsip.Type == BYE {
			if sendrecv == RECV {
				SendJSIPRes(jsip, 200)
			}
			session.state = INVITE_END
			return OK
		}
	case INVITE_ERR:
		if jsip.Type == ACK { // ERR ACK
			session.state = INVITE_END
			return IGNORE
		}
	}

	jstack.log.LogError("%s %s in %s", jsipDirect[sendrecv], JsipName(jsip),
		jsipInviteState[session.state])

	return ERROR
}

func jsipDefaultSession(session *JSIPSession, jsip *JSIP, sendrecv int) int {
	if jsip.Type == CANCEL {
		return OK
	}

	switch session.state {
	case DEFAULT_INIT:
		if jsip.Code != 0 {
			jstack.log.LogError("Recv response %s but session state is DEFAULT_INIT",
				JsipName(jsip))
			return ERROR
		}

		if session.typ != INVITE && session.typ != REGISTER &&
			session.typ != OPTIONS && session.typ != MESSAGE &&
			session.typ != SUBSCRIBE {

			jstack.log.LogError("Session not exist when process msg %s", JsipName(jsip))
			session.state = DEFAULT_REQ
			if sendrecv == RECV {
				SendJSIPRes(jsip, 481)
			}
			return ERROR
		}

		session.state = DEFAULT_REQ

	case DEFAULT_REQ:
		if jsip.Code == 0 {
			jstack.log.LogError("Recv request %s but session state is DEFAULT_REQ",
				JsipName(jsip))
			return ERROR
		}

		if jsip.Code >= 200 {
			session.state = DEFAULT_RESP
		}
	}

	return OK
}

func jsipSession(conn *JSIPConn, jsip *JSIP, sendrecv int) int {
	session := JsessGet(jsip.DialogueID)
	if session == nil {
		if jsip.Code != 0 {
			jstack.log.LogError("recv response but session is nil")
			return IGNORE
		}

		session = &JSIPSession{
			conn: conn,
			typ:  jsip.Type,
			req:  jsip,
			cseq: jsip.CSeq,
		}

		JsessSet(jsip.DialogueID, session)

		if sendrecv == RECV {
			session.uatype = UAS
		} else {
			session.uatype = UAC
		}

		switch session.typ {
		case INVITE:
			session.handler = jsipInviteSession
		default:
			session.handler = jsipDefaultSession
		}
	}

	switch session.handler(session, jsip, sendrecv) {
	case IGNORE:
		return IGNORE
	case ERROR:
		return ERROR
	}

	if jsip.Type == CANCEL && jsip.Code == 0 && sendrecv == RECV {
		relatedid, _ := jsip.RawMsg["RelatedID"]
		rid, _ := strconv.ParseUint(string(relatedid.(json.Number)), 10, 64)
		tid := transactionID(jsip, rid)
		cancelTrans := JtransGet(tid)
		// send CANCEL 200 and REQ 487
		SendJSIPRes(jsip, 200)
		SendJSIPRes(cancelTrans.req, 487)
	}

	if sendrecv == RECV {
		jstack.jsipHandle(jsip)
	} else {
		session.conn.sendq <- jsip.RawMsg
	}

	if session.typ == INVITE {
		if session.state == INVITE_END {
			JsessDel(session.req.DialogueID)
		}
	} else {
		if session.state == DEFAULT_RESP {
			JsessDel(session.req.DialogueID)
		}
	}

	return OK
}

func RecvJsonSIPMsg(conn *JSIPConn, data []byte) bool {
	// Syntax Layer
	jsip, err := jsipUnParser(data)
	if err != nil {
		jstack.log.LogError("UnParser Json sip message failed, msg: %s err: %v",
			string(data), err)
		return false
	}

	fmt.Println("Recv:", jsip)
	// Transaction Layer
	switch jsipTrasaction(jsip, RECV) {
	case ERROR:
		return false
	case IGNORE:
		return true
	}

	// Session Layer
	if jsipSession(conn, jsip, RECV) == ERROR {
		return false
	}

	return true
}

func SendJsonSIPMsg(conn *JSIPConn, jsip *JSIP) {
	// Syntax Layer
	j, err := jsipPrepared(jsip)
	if err != nil {
		jstack.log.LogError("Prepared Json sip message failed, err %v", err)
		return
	}

	fmt.Println("Send:", jsip)
	// Transaction Layer
	if jsipTrasaction(jsip, SEND) != OK {
		return
	}

	// Session Layer
	if jsipSession(conn, j, SEND) == ERROR {
		return
	}
}
