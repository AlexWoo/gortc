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
	"time"

	"github.com/go-ini/ini"
	"github.com/tidwall/gjson"
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
	TRANS_ERRRESP
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
	INVITE_CANCELLED
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
	DEFAULT_CANCELLED
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
	Body       string
	RawMsg     map[string]interface{}

	inner       bool
	conn        Conn
	Transaction *JSIPTrasaction
	Session     *JSIPSession
}

func JSIPMsgClone(req *JSIP, dlg string) *JSIP {
	msg := &JSIP{
		Type:       req.Type,
		Code:       req.Code,
		RequestURI: req.RequestURI,
		From:       req.From,
		To:         req.To,
		CSeq:       req.CSeq,
		DialogueID: dlg,
		Router:     req.Router,
		Body:       req.Body,
		RawMsg:     req.RawMsg,
	}

	return msg
}

func JSIPMsgRes(req *JSIP, code int) *JSIP {
	if req.Code != 0 {
		jstack.log.LogError("Cannot send response for response")
		return nil
	}

	resp := &JSIP{
		Type:       req.Type,
		Code:       code,
		From:       req.From,
		To:         req.To,
		CSeq:       req.CSeq,
		DialogueID: req.DialogueID,
		RawMsg:     make(map[string]interface{}),

		conn:        req.conn,
		Transaction: req.Transaction,
		Session:     req.Session,
	}

	return resp
}

func JSIPMsgAck(resp *JSIP) *JSIP {
	ack := &JSIP{
		Type:       ACK,
		RequestURI: resp.Transaction.req.RequestURI,
		From:       resp.From,
		To:         resp.To,
		DialogueID: resp.DialogueID,
		RawMsg:     make(map[string]interface{}),

		conn:    resp.conn,
		Session: resp.Session,
	}

	ack.SetInt("RelatedID", int64(resp.CSeq))

	if resp.Session == nil {
		ack.CSeq = resp.Transaction.cseq + 1
	}

	return ack
}

func JSIPMsgBye(session *JSIPSession) *JSIP {
	bye := &JSIP{
		Type:       BYE,
		RequestURI: session.req.RequestURI,
		From:       session.req.From,
		To:         session.req.To,
		DialogueID: session.req.DialogueID,
		RawMsg:     make(map[string]interface{}),

		conn:    session.conn,
		Session: session,
	}

	return bye
}

func JSIPMsgTerm(session *JSIPSession) *JSIP {
	if session.Type == INVITE {
		if session.State >= INVITE_ERR {
			return nil
		}

		if session.State >= INVITE_200 {
			bye := JSIPMsgBye(session)
			session.cseq++
			bye.CSeq = session.cseq

			return bye
		}

		if session.UAType == UAC {
			cancel := JSIPMsgCancel(session.req)
			session.cseq++
			cancel.CSeq = session.cseq

			return cancel
		} else {
			resp := JSIPMsgRes(session.req, 487)

			return resp
		}
	} else {
		if session.State >= TRANS_FINALRESP {
			return nil
		}

		if session.UAType == UAC {
			cancel := JSIPMsgCancel(session.req)
			session.cseq++
			cancel.CSeq = session.cseq

			return cancel
		} else {
			resp := JSIPMsgRes(session.req, 487)

			return resp
		}
	}
}

func JSIPMsgUpdate(session *JSIPSession) *JSIP {
	update := &JSIP{
		Type:       UPDATE,
		RequestURI: session.req.RequestURI,
		From:       session.req.From,
		To:         session.req.To,
		DialogueID: session.req.DialogueID,
		RawMsg:     make(map[string]interface{}),

		conn:    session.conn,
		Session: session,
	}

	return update
}

func JSIPMsgCancel(req *JSIP) *JSIP {
	cancel := &JSIP{
		Type:       CANCEL,
		RequestURI: req.RequestURI,
		From:       req.From,
		To:         req.To,
		DialogueID: req.DialogueID,
		RawMsg:     make(map[string]interface{}),

		conn:    req.conn,
		Session: req.Session,
	}

	cancel.SetInt("RelatedID", int64(req.CSeq))

	if req.Session == nil {
		cancel.CSeq = req.CSeq + 1
	}

	return cancel
}

func (jsip *JSIP) Name() string {
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

func (jsip *JSIP) Abstract() string {
	abstract := jsip.Name()
	if jsip.Code == 0 {
		abstract += " RequestURI: " + jsip.RequestURI
	}
	abstract += " From: " + jsip.From + " To: " + jsip.To + " CSeq: " +
		strconv.Itoa(int(jsip.CSeq)) + " DialogueID: " + jsip.DialogueID

	if len(jsip.Router) > 0 {
		abstract += " Router: " + jsip.Router[0]
		for i := 1; i < len(jsip.Router); i++ {
			abstract += "," + jsip.Router[0]
		}
	}

	relatedid, ok := jsip.GetInt("RelatedID")
	if ok {
		abstract += " RelatedID: " + strconv.Itoa(int(relatedid))
	}

	return abstract
}

func (jsip *JSIP) Detail() string {
	data, _ := json.Marshal(jsip.RawMsg)
	detail := jsip.Name() + ": " + string(data)

	return detail
}

func (jsip *JSIP) GetInt(header string) (int64, bool) {
	return getJsonInt64(jsip.RawMsg, header)
}

func (jsip *JSIP) SetInt(header string, value int64) {
	jsip.RawMsg[header] = value
}

func (jsip *JSIP) GetString(header string) (string, bool) {
	v, ok := jsip.RawMsg[header].(string)
	return v, ok
}

func (jsip *JSIP) SetString(header string, value string) {
	jsip.RawMsg[header] = value
}

type JSIPTrasaction struct {
	Type   int
	State  int
	UAType int
	req    *JSIP
	cseq   uint64
	tid    string

	timer *Timer

	conn Conn
}

func newJSIPTrans(tid string, jsip *JSIP, sendrecv int) *JSIPTrasaction {
	trans := &JSIPTrasaction{
		Type:  jsip.Type,
		State: TRANS_REQ,
		req:   jsip,
		cseq:  jsip.CSeq,
		tid:   tid,
	}

	if sendrecv == RECV {
		trans.UAType = UAS
		trans.conn = jsip.conn
	} else {
		trans.UAType = UAC
	}

	trans.timer = NewTimer(jstack.config.TransTimer, trans.timerHandle, nil)

	jstack.transactions[tid] = trans
	jsip.Transaction = trans

	return trans
}

func (trans *JSIPTrasaction) delete() {
	trans.timer.Stop()

	delete(jstack.transactions, trans.tid)
}

func (trans *JSIPTrasaction) timerHandle(t interface{}) {
	jstack.transTimeout <- trans
}

func (trans *JSIPTrasaction) timeout() {
	if trans.Type == CANCEL {
		trans.delete()
		return
	}

	if trans.State >= TRANS_FINALRESP {
		trans.delete()
		return
	}

	if trans.UAType == UAS {
		cancel := JSIPMsgCancel(trans.req)
		cancel.inner = true
		if cancel.Session != nil {
			cancel.Session.cseq++
			cancel.CSeq = cancel.Session.cseq
		}
		jstack.recvq_t <- cancel
	} else {
		resp := JSIPMsgRes(trans.req, 408)
		resp.inner = true
		jstack.recvq_t <- resp
	}
}

type JSIPSession struct {
	Type    int
	State   int
	UAType  int
	req     *JSIP
	dlg     string
	cseq    uint64
	handler func(session *JSIPSession, jsip *JSIP, sendrecv int) int

	expire    time.Duration
	sessTimer *Timer
	err       bool

	conn Conn
}

func newJSIPSess(jsip *JSIP, sendrecv int) *JSIPSession {
	session := &JSIPSession{
		Type: jsip.Type,
		req:  jsip,
		dlg:  jsip.DialogueID,
	}

	if sendrecv == RECV {
		session.UAType = UAS
		session.conn = jsip.conn
	} else {
		session.UAType = UAC
	}

	jstack.sessions[jsip.DialogueID] = session
	jsip.Session = session

	switch session.Type {
	case INVITE:
		session.handler = jstack.jsipInviteSession

		expire := jstack.config.SessionTimer
		exp, ok := jsip.GetInt("Expire")
		if ok && exp >= 60 {
			expire = time.Duration(exp) * time.Second
		}

		if session.UAType == UAS {
			session.expire = expire
		} else {
			session.expire = expire/2 - jstack.config.TransTimer
		}
	default:
		session.handler = jstack.jsipDefaultSession
	}

	return session
}

func (sess *JSIPSession) delete() {
	if sess.sessTimer != nil {
		sess.sessTimer.Stop()
	}

	delete(jstack.sessions, sess.dlg)
}

func (sess *JSIPSession) errorHandle() {
	if sess.Type != INVITE {
		return
	}

	if sess.State >= INVITE_ERR {
		return
	}

	if sess.State >= INVITE_200 {
		// Send BYE to application layer
		bye := JSIPMsgBye(sess)
		bye.inner = true
		sess.cseq++
		bye.CSeq = sess.cseq
		jstack.recvq_t <- bye

		return
	}

	// sess.State < INVITE_200
	if sess.UAType == UAS {
		// Send CANCEL to application layer
		cancel := JSIPMsgCancel(sess.req)
		cancel.inner = true
		sess.cseq++
		cancel.CSeq = sess.cseq
		jstack.recvq_t <- cancel
	} else {
		// Send INVITE_408 to application layer
		resp := JSIPMsgRes(sess.req, 408)
		resp.inner = true
		jstack.recvq_t <- resp
	}
}

func (sess *JSIPSession) timerHandle(t interface{}) {
	jstack.sessTimeout <- sess
}

func (sess *JSIPSession) sessionTimer() {
	if sess.UAType == UAS {
		sess.errorHandle()
	} else {
		if sess.err {
			sess.errorHandle()
		} else {
			update := JSIPMsgUpdate(sess)
			sess.cseq++
			update.CSeq = sess.cseq
			jstack.sendq_t <- update
			sess.err = true
		}
	}
}

type JSIPConfig struct {
	Location     string `default:"rtc"`
	Realm        string
	Timeout      time.Duration `default:"1s"`
	Qsize        Size_t        `default:"1k"`
	TransTimer   time.Duration `default:"5s"`
	PRTimer      time.Duration `default:"60s"`
	SessionLayer bool          `default:"true"`
	SessionTimer time.Duration `default:"600s"`
}

type JSIPStack struct {
	confPath string
	config   *JSIPConfig
	jsipC    chan *JSIP
	log      *Log

	recvq_t      chan *JSIP
	sendq_t      chan *JSIP
	recvq_s      chan *JSIP
	sendq_s      chan *JSIP
	transTimeout chan *JSIPTrasaction
	sessTimeout  chan *JSIPSession

	sessions     map[string]*JSIPSession
	transactions map[string]*JSIPTrasaction
}

var jstack *JSIPStack

func (stack *JSIPStack) loadConfig() bool {
	stack.config = new(JSIPConfig)

	f, err := ini.Load(stack.confPath)
	if err != nil {
		jstack.log.LogError("Load config file %s error: %v",
			stack.confPath, err)
		return false
	}

	return Config(f, "JSIPStack", stack.config)
}

func (stack *JSIPStack) transactionID(jsip *JSIP, cseq uint64) string {
	return jsip.DialogueID + "_" + strconv.FormatUint(cseq, 10)
}

func (stack *JSIPStack) parseUri(uri string) (string, string) {
	var userWithHost, hostWithPort string

	ss := strings.Split(uri, ";")
	uri = ss[0]

	ss = strings.Split(uri, "@")
	if len(ss) == 1 {
		hostWithPort = ss[0]
	} else {
		hostWithPort = ss[1]
	}

	ss = strings.Split(uri, ":")
	userWithHost = ss[0]

	return userWithHost, hostWithPort
}

func (stack *JSIPStack) connect(uri string) *WSConn {
	userWithHost, hostWithPort := stack.parseUri(uri)
	if userWithHost == "" {
		return nil
	}

	url := "ws://" + hostWithPort + jstack.Location() + "?userid=" +
		jstack.Realm()
	conn := NewWSConn(userWithHost, url, UAC, jstack.Timeout(), jstack.Qsize(),
		RecvMsg)

	return conn
}

// Syntax Check
func getJsonInt(j map[string]interface{}, key string) (int, bool) {
	value, ok := j[key].(float64)
	if !ok {
		return 0, false
	}

	return int(value), true
}

func getJsonInt64(j map[string]interface{}, key string) (int64, bool) {
	value, ok := j[key].(float64)
	if !ok {
		return 0, false
	}

	return int64(value), true
}

func getJsonUint(j map[string]interface{}, key string) (uint, bool) {
	value, ok := j[key].(float64)
	if !ok {
		return 0, false
	}

	return uint(value), true
}

func getJsonUint64(j map[string]interface{}, key string) (uint64, bool) {
	value, ok := j[key].(float64)
	if !ok {
		return 0, false
	}

	return uint64(value), true
}

func (stack *JSIPStack) jsipUnParser(data []byte) (*JSIP, error) {
	j, ok := gjson.ParseBytes(data).Value().(map[string]interface{})
	if !ok {
		return nil, errors.New("data is not json object")
	}

	jsip := &JSIP{
		RawMsg: j,
	}

	typ, ok := j["Type"].(string)
	if !ok {
		return nil, errors.New("Type error")
	}

	if typ == "RESPONSE" {
		jsip.Code, ok = getJsonInt(j, "Code")
		if !ok {
			return nil, errors.New("Code error")
		}

		if jsip.Code < 100 || jsip.Code > 699 {
			return nil, fmt.Errorf("Unexpected status code %d", jsip.Code)
		}
	} else {
		jsip.Type = jsipReqUnparse[typ]
		if jsip.Type == UNKNOWN {
			return nil, errors.New("Unknown Type")
		}

		jsip.RequestURI, ok = j["Request-URI"].(string)
		if !ok {
			return nil, errors.New("Request-URI error")
		}
	}

	jsip.From, ok = j["From"].(string)
	if !ok {
		return nil, errors.New("From error")
	}

	jsip.To, ok = j["To"].(string)
	if !ok {
		return nil, errors.New("To error")
	}

	jsip.DialogueID, ok = j["DialogueID"].(string)
	if !ok {
		return nil, errors.New("DialogueID error")
	}

	jsip.CSeq, ok = getJsonUint64(j, "CSeq")
	if !ok {
		return nil, errors.New("CSeq error")
	}

	routers, ok := j["Router"].(string)
	if ok {
		jsip.Router = strings.Split(routers, ",")
		for i := 0; i < len(jsip.Router); i++ {
			jsip.Router[i] = strings.TrimSpace(jsip.Router[i])
		}
	}

	body, ok := j["Body"]
	if ok {
		b, _ := json.Marshal(body)
		jsip.Body = string(b)
	}

	return jsip, nil
}

func (stack *JSIPStack) jsipParser(jsip *JSIP) *JSIP {
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

	if jsip.Body != "" {
		jsip.RawMsg["Body"] = jsip.Body
	}

	return jsip
}

func (stack *JSIPStack) jsipPrepared(jsip *JSIP) (*JSIP, error) {
	if jsip.DialogueID == "" {
		return nil, errors.New("DialogueID not set")
	}

	if jsipReqParse[jsip.Type] == "" {
		return nil, errors.New("Unknown message type")
	}

	if jsip.Code != 0 && (jsip.Code < 100 || jsip.Code > 699) {
		return nil, fmt.Errorf("Unknown response %s", jsip.Name())
	}

	if jsip.From == "" {
		return nil, errors.New("From not set")
	}

	if jsip.To == "" {
		return nil, errors.New("To not set")
	}

	if jsip.Code == 0 {
		if jsip.RequestURI == "" {
			return nil, errors.New("RequestURI not set in JSIP Request")
		}
	} else {
		if jsip.CSeq == 0 {
			return nil, errors.New("CSeq not set in JSIP Response")
		}
	}

	if jsip.RawMsg == nil {
		jsip.RawMsg = make(map[string]interface{})
	}

	return jsip, nil
}

// Transaction Layer
func (stack *JSIPStack) jsipTransaction(jsip *JSIP, sendrecv int) int {
	tid := stack.transactionID(jsip, jsip.CSeq)
	trans := stack.transactions[tid]
	jsip.Transaction = trans

	if trans == nil { // Request
		if jsip.Code != 0 {
			stack.log.LogError("process %s but trans is nil", jsip.Name())
			return ERROR
		}

		trans = newJSIPTrans(tid, jsip, sendrecv)

		if jsip.Type == ACK {
			trans.delete()

			rid, ok := jsip.GetInt("RelatedID")
			if !ok {
				stack.log.LogInfo("ACK miss RelatedID")
				return IGNORE
			}

			ackid := stack.transactionID(jsip, uint64(rid))
			ackTrans := stack.transactions[ackid]
			if ackTrans == nil {
				stack.log.LogInfo("Transaction INVITE not exist")
				return IGNORE
			}

			if ackTrans.State < TRANS_FINALRESP {
				stack.log.LogError("Recv ACK but not receive final response")
				return IGNORE
			}

			if ackTrans.UAType == UAS && sendrecv == SEND ||
				ackTrans.UAType == UAC && sendrecv == RECV {

				stack.log.LogError("ACK direct is not same as INVITE")
				return IGNORE
			}

			ackTrans.delete()

			if ackTrans.State == TRANS_ERRRESP && sendrecv == RECV {
				return IGNORE
			}
		}

		if jsip.Type == CANCEL {
			rid, ok := jsip.GetInt("RelatedID")
			if !ok {
				trans.delete()

				stack.log.LogInfo("CANCEL miss RelatedID")
				return IGNORE
			}

			cancelid := stack.transactionID(jsip, uint64(rid))
			cancelTrans := stack.transactions[cancelid]
			if cancelTrans == nil {
				trans.delete()

				stack.log.LogInfo("Transaction Cancelled not exist")
				return IGNORE
			}

			if cancelTrans.State >= TRANS_FINALRESP {
				trans.delete()

				stack.log.LogInfo("Transaction in finalize response, cannot cancel")
				return IGNORE
			}

			if cancelTrans.UAType == UAS && sendrecv == SEND ||
				cancelTrans.UAType == UAC && sendrecv == RECV {

				trans.delete()

				stack.log.LogError("CANCEL direct is not same as Request")
				return IGNORE
			}

			if sendrecv == RECV {
				// Send CANCLE 200
				if jsip.inner {
					stack.sendq_t <- JSIPMsgRes(cancelTrans.req, 408)
					trans.delete()
				} else {
					stack.sendq_t <- JSIPMsgRes(jsip, 200)
					stack.sendq_t <- JSIPMsgRes(cancelTrans.req, 487)
				}
			}

			if cancelTrans.Type != INVITE && cancelTrans.Type != REGISTER &&
				cancelTrans.Type != OPTIONS && cancelTrans.Type != MESSAGE &&
				cancelTrans.Type != SUBSCRIBE {

				return IGNORE
			}
		}

		if jsip.Type == BYE {
			if sendrecv == RECV {
				// Send BYE 200
				if jsip.inner {
					trans.delete()
				} else {
					stack.sendq_t <- JSIPMsgRes(jsip, 200)
				}
			}
		}

		return OK
	}

	if jsip.Code == 0 {
		stack.log.LogError("process %s but trans exist", jsip.Name())
		return IGNORE
	}

	// Response
	if trans.UAType == UAS && sendrecv == RECV ||
		trans.UAType == UAC && sendrecv == SEND {

		stack.log.LogError("Response direct is same as Request direct")
		return IGNORE
	}

	jsip.Type = trans.Type

	if jsip.Code == 100 {
		if trans.State > TRANS_TRYING {
			stack.log.LogError("process 100 Trying but state is %d", trans.State)
			return IGNORE
		}

		trans.State = TRANS_TRYING

		return IGNORE
	}

	if jsip.Code < 200 && jsip.Code > 100 {
		if trans.State > TRANS_PR {
			stack.log.LogError("process %s but state is %d", jsip.Name(),
				trans.State)
			return IGNORE
		}

		trans.timer.Reset(stack.config.PRTimer)
		trans.State = TRANS_PR

		return OK
	}

	if trans.State >= TRANS_FINALRESP {
		stack.log.LogError("process %s but state is %d", jsip.Name(),
			trans.State)
		return IGNORE
	}

	if jsip.Code >= 300 {
		trans.State = TRANS_ERRRESP
	} else {
		trans.State = TRANS_FINALRESP
	}

	if trans.Type != INVITE {
		trans.delete()
	} else {
		if jsip.Code >= 300 && sendrecv == RECV {
			if jsip.inner {
				trans.delete()
				return OK
			}
			// Send ACK for INVITE 3XX 4XX 5XX 6XX Response
			stack.sendq_t <- JSIPMsgAck(jsip)
		}
		// Wait for ACK
		trans.timer.Reset(stack.config.TransTimer)
	}

	if trans.Type == CANCEL && sendrecv == RECV {
		// Ignore CANCEL 200 received
		return IGNORE
	}

	if trans.Type == BYE && sendrecv == RECV {
		// Ignore BYE 200 received
		return IGNORE
	}

	return OK
}

// Session Layer
func (stack *JSIPStack) jsipInviteSession(session *JSIPSession, jsip *JSIP,
	sendrecv int) int {

	if jsip.Type == INFO {
		return OK
	}

	if jsip.Type == CANCEL && jsip.Code > 0 {
		return IGNORE
	}

	if jsip.Type == CANCEL {
		if session.State >= INVITE_200 && session.State < INVITE_ERR {
			goto err
		}

		session.State = INVITE_CANCELLED

		return OK
	}

	if jsip.Type == UPDATE &&
		session.State >= INVITE_200 && session.State < INVITE_ERR {

		if jsip.Code == 0 {
			if sendrecv == RECV && session.UAType == UAS {
				stack.sendq_t <- JSIPMsgRes(jsip, 200)
				session.sessTimer.Reset(session.expire)

				return IGNORE
			}
		}

		if jsip.Code == 200 {
			if sendrecv == RECV && session.UAType == UAC {
				session.sessTimer.Reset(session.expire)
				session.err = false

				return IGNORE
			}

		}

		goto err
	}

	if jsip.Type == BYE {
		if session.State < INVITE_200 {
			goto err
		}

		session.State = INVITE_END

		return OK
	}

	switch session.State {
	case INVITE_INIT:
		if jsip.Type == INVITE && jsip.Code == 0 {
			session.State = INVITE_REQ
			return OK
		}
	case INVITE_REQ:
		switch jsip.Type {
		case INVITE:
			switch {
			case jsip.Code == 100:
				return OK
			case jsip.Code < 200 && jsip.Code > 100:
				session.State = INVITE_18X
				return OK
			case jsip.Code == 200:
				session.State = INVITE_200
				session.sessTimer = NewTimer(session.expire,
					session.timerHandle, nil)
				return OK
			case jsip.Code >= 300:
				session.State = INVITE_ERR
				return OK
			}
		}
	case INVITE_18X:
		switch jsip.Type {
		case INVITE:
			switch {
			case jsip.Code < 200 && jsip.Code > 100:
				return OK
			case jsip.Code == 200:
				session.State = INVITE_200
				session.sessTimer = NewTimer(session.expire,
					session.timerHandle, nil)
				return OK
			case jsip.Code >= 300:
				session.State = INVITE_ERR
				return OK
			}
		case PRACK:
			if jsip.Code == 0 && sendrecv == session.UAType {
				session.State = INVITE_PRACK
				return OK
			}
		case UPDATE:
			if jsip.Code == 0 {
				session.State = INVITE_UPDATE
				return OK
			}
		}
	case INVITE_PRACK:
		if jsip.Code == 200 && jsip.Type == PRACK {
			session.State = INVITE_18X
			return OK
		}
	case INVITE_UPDATE:
		if jsip.Code == 200 && jsip.Type == UPDATE {
			session.State = INVITE_18X
			return OK
		}
	case INVITE_200:
		if jsip.Type == ACK {
			session.State = INVITE_ACK
			return OK
		}
	case INVITE_ACK:
		switch {
		case jsip.Type == INVITE:
			if jsip.Code == 0 {
				session.State = INVITE_REINV
				return OK
			}
		case jsip.Type == INFO: // INFO and INFO 200
			return OK
		}
	case INVITE_REINV:
		if jsip.Code == 200 && jsip.Type == INVITE {
			session.State = INVITE_RE200
			return OK
		}
	case INVITE_RE200:
		if jsip.Type == ACK {
			session.State = INVITE_ACK
			return OK
		}
	}

err:
	if session.State >= INVITE_200 && session.State < INVITE_ERR {
		session.State = INVITE_ACK
	} else if session.State < INVITE_200 {
		session.State = INVITE_18X
	}

	stack.log.LogError("%s %s in %s", jsipDirect[sendrecv], jsip.Name(),
		jsipInviteState[session.State])

	return ERROR
}

func (stack *JSIPStack) jsipDefaultSession(session *JSIPSession, jsip *JSIP,
	sendrecv int) int {

	if jsip.Type == CANCEL && jsip.Code > 0 {
		return IGNORE
	}

	switch session.State {
	case DEFAULT_INIT:
		if jsip.Code != 0 {
			stack.log.LogError("Recv response %s but session state is DEFAULT_INIT",
				jsip.Name())
			return ERROR
		}

		if session.Type == CANCEL {
			stack.log.LogError("Session not exist when process msg %s", jsip.Name())
			session.State = DEFAULT_RESP

			resp := JSIPMsgRes(jsip, 481)
			if sendrecv == RECV {
				stack.sendq_t <- resp
			} else {
				stack.log.LogDebug("Recv: %s", resp.Abstract())
				stack.jsipC <- resp
			}

			return ERROR
		} else if session.Type != INVITE && session.Type != REGISTER &&
			session.Type != OPTIONS && session.Type != MESSAGE &&
			session.Type != SUBSCRIBE {

			stack.log.LogError("Session not exist when process msg %s", jsip.Name())
			session.State = DEFAULT_REQ

			resp := JSIPMsgRes(jsip, 481)
			if sendrecv == RECV {
				stack.sendq_s <- resp
			} else {
				stack.jsipC <- resp
			}
			return ERROR
		}

		session.State = DEFAULT_REQ

	case DEFAULT_REQ:
		if jsip.Code == 0 {
			if jsip.Type == CANCEL {
				session.State = DEFAULT_CANCELLED
				return OK
			}

			stack.log.LogError("Recv request %s but session state is DEFAULT_REQ",
				jsip.Name())

			resp := JSIPMsgRes(jsip, 400)
			if sendrecv == RECV {
				stack.sendq_t <- resp
			} else {
				stack.recvq_s <- resp
			}

			return ERROR
		}

		if jsip.Code >= 200 && session.req.CSeq == jsip.CSeq {
			session.State = DEFAULT_RESP
		}
	}

	return OK
}

func (stack *JSIPStack) jsipSession(jsip *JSIP, sendrecv int) int {
	if !stack.config.SessionLayer {
		return OK
	}

	session := stack.sessions[jsip.DialogueID]
	jsip.Session = session

	if session == nil {
		if jsip.Code != 0 {
			stack.log.LogError("recv response but session is nil")
			return IGNORE
		}

		session = newJSIPSess(jsip, sendrecv)
	}

	if jsip.Code == 0 {
		if sendrecv == RECV {
			session.cseq = jsip.CSeq
		} else {
			session.cseq++
			jsip.CSeq = session.cseq
		}
	}

	ret := session.handler(session, jsip, sendrecv)
	if ret == ERROR {
		session.errorHandle()
	}

	if session.Type == INVITE {
		if session.State >= INVITE_ERR {
			session.delete()
		}
	} else {
		if session.State >= DEFAULT_RESP {
			session.delete()
		}
	}

	return ret
}

func (stack *JSIPStack) recvJSIPMsg_t(jsip *JSIP) {
	// Transaction Layer
	stack.log.LogDebug("Recv[Transaction]: %s", jsip.Abstract())
	ret := stack.jsipTransaction(jsip, RECV)
	if ret == ERROR {
		return
	} else if ret == IGNORE {
		return
	}

	stack.recvq_s <- jsip
}

func (stack *JSIPStack) recvJSIPMsg_s(jsip *JSIP) {
	// Session Layer
	stack.log.LogDebug("Recv[Session]: %s", jsip.Abstract())
	ret := stack.jsipSession(jsip, RECV)
	if ret == ERROR {
		return
	} else if ret == IGNORE {
		return
	}

	stack.log.LogDebug("Recv: %s", jsip.Abstract())
	stack.jsipC <- jsip
}

func (stack *JSIPStack) sendJSIPMsg_s(jsip *JSIP) {
	// Session Layer
	stack.log.LogDebug("Send[Session]: %s", jsip.Abstract())
	ret := stack.jsipSession(jsip, SEND)
	if ret == ERROR {
		return
	} else if ret == IGNORE {
		return
	}

	stack.sendq_t <- jsip
}

func (stack *JSIPStack) sendJSIPMsg_t(jsip *JSIP) {
	// Transaction Layer
	stack.log.LogDebug("Send[Transaction]: %s", jsip.Abstract())
	ret := stack.jsipTransaction(jsip, SEND)
	if ret == ERROR {
		return
	} else if ret == IGNORE {
		return
	}

	if jsip.Transaction.conn != nil {
		jsip.conn = jsip.Transaction.conn
	} else {
		if jsip.Session != nil {
			jsip.conn = jsip.Session.conn
		}
	}

	if jsip.conn == nil {
		var conn *WSConn
		if len(jsip.Router) > 0 {
			conn = stack.connect(jsip.Router[0])
		} else {
			conn = stack.connect(jsip.RequestURI)
		}

		if conn == nil {
			//TODO Error
			return
		}

		jsip.conn = conn
		jsip.Transaction.conn = conn
		if jsip.Session != nil {
			jsip.Session.conn = conn
		}
	}

	jsip = stack.jsipParser(jsip)

	data, _ := json.Marshal(jsip.RawMsg)
	stack.log.LogDebug("Send[Raw] %s", string(data))
	jsip.conn.Send(data)
}

func (stack *JSIPStack) run() {
	for {
		select {
		case jsip := <-stack.recvq_t:
			stack.recvJSIPMsg_t(jsip)
		case jsip := <-stack.sendq_t:
			stack.sendJSIPMsg_t(jsip)
		case jsip := <-stack.recvq_s:
			stack.recvJSIPMsg_s(jsip)
		case jsip := <-stack.sendq_s:
			stack.sendJSIPMsg_s(jsip)
		case trans := <-stack.transTimeout:
			trans.timeout()
		case sess := <-stack.sessTimeout:
			sess.sessionTimer()
		}
	}
}

// Init JSIP Stack
func InitJSIPStack(jsipC chan *JSIP, log *Log, rtcpath string) *JSIPStack {
	jstack = &JSIPStack{
		confPath:     rtcpath + "/conf/gortc.ini",
		jsipC:        jsipC,
		log:          log,
		sessions:     make(map[string]*JSIPSession),
		transactions: make(map[string]*JSIPTrasaction),
	}

	if !jstack.loadConfig() {
		return nil
	}

	if jstack.config.Realm == "" {
		jstack.log.LogError("JSIPStack Realm not configured")
		return nil
	}

	jstack.recvq_t = make(chan *JSIP, jstack.Qsize())
	jstack.sendq_t = make(chan *JSIP, jstack.Qsize())
	jstack.recvq_s = make(chan *JSIP, jstack.Qsize())
	jstack.sendq_s = make(chan *JSIP, jstack.Qsize())
	jstack.transTimeout = make(chan *JSIPTrasaction)
	jstack.sessTimeout = make(chan *JSIPSession)

	go jstack.run()

	return jstack
}

// JStack Config
func Realm() string {
	return jstack.config.Realm
}

func (stack *JSIPStack) Location() string {
	return stack.config.Location
}

func (stack *JSIPStack) Realm() string {
	return stack.config.Realm
}

func (stack *JSIPStack) Timeout() time.Duration {
	return stack.config.Timeout
}

func (stack *JSIPStack) Qsize() uint64 {
	return uint64(stack.config.Qsize)
}

func RecvMsg(conn Conn, data []byte) {
	jstack.log.LogDebug("Recv[Raw]: %s", string(data))
	jsip, err := jstack.jsipUnParser(data)
	if err != nil {
		jstack.log.LogError("jsipUnParser err %s", err)
		return
	}

	jsip.conn = conn

	jstack.recvq_t <- jsip
}

func SendMsg(jsip *JSIP) {
	jstack.log.LogDebug("Send: %s", jsip.Abstract())
	jsip, err := jstack.jsipPrepared(jsip)
	if err != nil {
		jstack.log.LogError("jsipPrepared err %s", err)
		return
	}

	jstack.sendq_s <- jsip
}
