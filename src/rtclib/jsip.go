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

	"github.com/alexwoo/golib"
	"github.com/tidwall/gjson"
)

// return value
const (
	ERROR = iota
	OK
	IGNORE
	DONE
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
	NOTIFY
	TERM
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
	"NOTIFY":    NOTIFY,
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
	NOTIFY:    "NOTIFY",
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

func copySlice(src []interface{}) []interface{} {
	dst := make([]interface{}, 0, len(src))

	for _, v := range src {
		if d, ok := v.(map[string]interface{}); ok {
			dst = append(dst, copyMap(d))
		} else if d, ok := v.([]interface{}); ok {
			dst = append(dst, copySlice(d))
		} else {
			dst = append(dst, v)
		}
	}

	return dst
}

func copyMap(src map[string]interface{}) map[string]interface{} {
	dst := make(map[string]interface{})

	for k, v := range src {
		if d, ok := v.(map[string]interface{}); ok {
			dst[k] = copyMap(d)
		} else if d, ok := v.([]interface{}); ok {
			dst[k] = copySlice(d)
		} else {
			dst[k] = v
		}
	}

	return dst
}

func copyBody(src interface{}) interface{} {
	if s, ok := src.(string); ok {
		return s
	}

	if s, ok := src.(map[string]interface{}); ok {
		return copyMap(s)
	}

	if s, ok := src.([]interface{}); ok {
		return copySlice(s)
	}

	return nil
}

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

// Parse a JSIP URI, format as [[prefix:]user@]host[:port][;para1=value][;para2]
func ParseJSIPUri(uri string) (*JSIPUri, error) {
	if uri == "" {
		return nil, errors.New("Null uri")
	}

	jsipUri := &JSIPUri{
		Paras: make(map[string]interface{}),
	}

	ss := strings.Split(uri, ";")
	if len(ss) >= 2 { // Has paras
		paras := ss[1:]
		for _, para := range paras {
			pp := strings.SplitN(para, "=", 2)
			if pp[0] == "" {
				return nil, errors.New("Null para")
			}

			if len(pp) == 1 {
				jsipUri.Paras[pp[0]] = true
			} else {
				jsipUri.Paras[pp[0]] = pp[1]
			}
		}
	}

	hostPart := ss[0]
	ss = strings.Split(hostPart, "@")
	if len(ss) == 1 {
		jsipUri.HostWithPort = ss[0]
	} else if len(ss) == 2 {
		jsipUri.UserWithPrefix = ss[0]
		jsipUri.HostWithPort = ss[1]
	} else {
		return nil, errors.New("Too many '@' in host part")
	}

	ss = strings.Split(jsipUri.HostWithPort, ":")
	jsipUri.Host = ss[0]
	if len(ss) == 2 { // Has port
		port, err := strconv.ParseUint(ss[1], 10, 16)
		if err != nil {
			return nil, err
		}
		jsipUri.Port = uint16(port)
	} else if len(ss) > 2 {
		return nil, errors.New("HostWithPort format error")
	}

	if jsipUri.UserWithPrefix != "" { // Has UserWithPrefix
		ss = strings.Split(jsipUri.UserWithPrefix, ":")
		if len(ss) == 1 {
			jsipUri.User = ss[0]
		} else if len(ss) == 2 {
			jsipUri.Prefix = ss[0]
			jsipUri.User = ss[1]
		} else {
			return nil, errors.New("UserWithPrefix format error")
		}
	}

	if jsipUri.UserWithPrefix != "" {
		jsipUri.UserWithHost = jsipUri.UserWithPrefix + "@" + jsipUri.Host
	} else {
		jsipUri.UserWithHost = jsipUri.Host
	}

	return jsipUri, nil
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

	inner       bool
	conn        golib.Conn
	Transaction *JSIPTrasaction
	Session     *JSIPSession
	Userid      string
}

// Get jsip name
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

// Get jsip abstract description
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

// Get jsip detail content
func (jsip *JSIP) Detail() string {
	data, _ := json.Marshal(jsip.RawMsg)
	detail := jsip.Name() + ": " + string(data)

	return detail
}

// Get jsip header with integer
func (jsip *JSIP) GetInt(header string) (int64, bool) {
	return getJsonInt64(jsip.RawMsg, header)
}

// Set jsip header with integer
func (jsip *JSIP) SetInt(header string, value int64) {
	jsip.RawMsg[header] = value
}

// Get jsip header with type string
func (jsip *JSIP) GetString(header string) (string, bool) {
	v, ok := jsip.RawMsg[header].(string)
	return v, ok
}

// Set jsip header with type string
func (jsip *JSIP) SetString(header string, value string) {
	jsip.RawMsg[header] = value
}

// for log ctx

func (jsip *JSIP) Prefix() string {
	return "[jstack]"
}

func (jsip *JSIP) Suffix() string {
	suf := ", " + jsip.Name()

	if jsip.conn != nil {
		suf += jsip.conn.Suffix()
	}

	return suf
}

func (jsip *JSIP) LogLevel() int {
	return jstack.logLevel
}

type JSIPTrasaction struct {
	Type   int
	State  int
	UAType int
	req    *JSIP
	cseq   uint64
	tid    string

	timer *golib.Timer

	conn golib.Conn
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

	trans.timer = golib.NewTimer(jstack.dconfig.TransTimer,
		trans.timerHandle, nil)

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
	sessTimer *golib.Timer
	err       bool

	conn golib.Conn
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

		expire := jstack.dconfig.SessionTimer
		exp, ok := jsip.GetInt("Expire")
		if ok && exp >= 60 {
			expire = time.Duration(exp) * time.Second
		}

		if session.UAType == UAS {
			session.expire = expire
		} else {
			session.expire = expire/2 - jstack.dconfig.TransTimer
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

type jsipConfig struct {
	Qsize uint64 `default:"1024"`
}

type jsipDConfig struct {
	Realm        string
	Location     string        `default:"/rtc"`
	ConnTimeout  time.Duration `default:"3s"`
	Retry        int64         `default:"10"`
	TransTimer   time.Duration `default:"5s"`
	PRTimer      time.Duration `default:"60s"`
	SessionLayer bool          `default:"true"`
	SessionTimer time.Duration `default:"600s"`
	TermNotify   bool          `default:"false"`
}

type JSIPStack struct {
	config   *jsipConfig
	dconfig  *jsipDConfig
	log      *golib.Log
	logLevel int

	jsipC        chan *JSIP
	recvq_t      chan *JSIP
	sendq_t      chan *JSIP
	recvq_s      chan *JSIP
	sendq_s      chan *JSIP
	transTimeout chan *JSIPTrasaction
	sessTimeout  chan *JSIPSession
	close        chan bool

	sessions     map[string]*JSIPSession
	transactions map[string]*JSIPTrasaction
}

var jstack *JSIPStack

func JStackInstance() *JSIPStack {
	if jstack != nil {
		return jstack
	}

	jstack = &JSIPStack{
		transTimeout: make(chan *JSIPTrasaction),
		sessTimeout:  make(chan *JSIPSession),
		close:        make(chan bool),
		sessions:     make(map[string]*JSIPSession),
		transactions: make(map[string]*JSIPTrasaction),
	}

	return jstack
}

func (m *JSIPStack) loadDConfig() error {
	confPath := FullPath("conf/gortc.ini")

	config := &jsipDConfig{}
	err := golib.ConfigFile(confPath, "JSIPStack", config)
	if err != nil {
		return fmt.Errorf("Parse dconfig %s Failed, %s", confPath, err)
	}
	m.dconfig = config

	return nil
}

func (m *JSIPStack) loadConfig() error {
	confPath := FullPath("conf/gortc.ini")

	config := &jsipConfig{}
	err := golib.ConfigFile(confPath, "JSIPStack", config)
	if err != nil {
		return fmt.Errorf("Parse config %s Failed, %s", confPath, err)
	}
	m.config = config

	return nil
}

func (m *JSIPStack) SetLog(log *golib.Log, logLevel int) {
	m.log = log
	m.logLevel = logLevel
}

// for module interface

func (m *JSIPStack) PreInit() error {
	if err := m.loadConfig(); err != nil {
		return err
	}

	if err := m.loadDConfig(); err != nil {
		return err
	}

	return nil
}

func (m *JSIPStack) Init() error {
	if m.dconfig.Realm == "" {
		return fmt.Errorf("JSIP Stack Realm not configured")
	}

	if m.log == nil {
		return fmt.Errorf("JSIP Stack log not set")
	}

	m.jsipC = make(chan *JSIP, m.config.Qsize)
	m.recvq_t = make(chan *JSIP, m.config.Qsize)
	m.sendq_t = make(chan *JSIP, m.config.Qsize)
	m.recvq_s = make(chan *JSIP, m.config.Qsize)
	m.sendq_s = make(chan *JSIP, m.config.Qsize)

	return nil
}

func (m *JSIPStack) PreMainloop() error {
	return nil
}

func (m *JSIPStack) Mainloop() {
	for {
		select {
		case jsip := <-m.recvq_t:
			m.recvJSIPMsg_t(jsip)
		case jsip := <-m.sendq_t:
			m.sendJSIPMsg_t(jsip)
		case jsip := <-m.recvq_s:
			m.recvJSIPMsg_s(jsip)
		case jsip := <-m.sendq_s:
			m.sendJSIPMsg_s(jsip)
		case trans := <-m.transTimeout:
			trans.timeout()
		case sess := <-m.sessTimeout:
			sess.sessionTimer()
		case <-m.close:
			return
		}
	}
}

func (m *JSIPStack) Reload() error {
	if err := m.loadDConfig(); err != nil {
		return err
	}

	return nil
}

func (m *JSIPStack) Reopen() error {
	return nil
}

func (m *JSIPStack) Exit() {
	m.close <- true
}

// for log ctx

func (m *JSIPStack) Prefix() string {
	return "[jstack]"
}

func (m *JSIPStack) Suffix() string {
	return ""
}

func (m *JSIPStack) LogLevel() int {
	return m.logLevel
}

// internal interface

func (m *JSIPStack) transactionID(jsip *JSIP, cseq uint64) string {
	return jsip.DialogueID + "_" + strconv.FormatUint(cseq, 10)
}

func (m *JSIPStack) connect(uri string, userid string) *golib.WSConn {
	jsipUri, err := ParseJSIPUri(uri)
	if err != nil {
		m.log.LogError(m, "Parse uri %s error: %v", uri, err)
		return nil
	}

	if userid == "" {
		userid = m.dconfig.Realm
	}

	url := "ws://" + jsipUri.HostWithPort + m.dconfig.Location + "?userid=" +
		userid
	conn := golib.NewWSClient(jsipUri.UserWithHost, url, m.dconfig.ConnTimeout,
		int(m.dconfig.Retry), m.config.Qsize, RecvMsg, m.log, m.logLevel)

	return conn
}

// Syntax Check
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

	r := gjson.GetBytes(data, "Body")
	if r.Exists() {
		jsip.Body = r.Value()
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

	if jsip.Body != nil {
		jsip.RawMsg["Body"] = jsip.Body
	}

	return jsip
}

func (stack *JSIPStack) jsipPrepared(jsip *JSIP) (*JSIP, error) {
	if jsip.DialogueID == "" {
		return nil, errors.New("DialogueID not set")
	}

	if jsip.Type == TERM {
		return jsip, nil
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
func (m *JSIPStack) jsipTransaction(jsip *JSIP, sendrecv int) int {
	tid := m.transactionID(jsip, jsip.CSeq)
	trans := m.transactions[tid]
	jsip.Transaction = trans

	if trans == nil { // Request
		if jsip.Code != 0 {
			m.log.LogError(jsip, "process %s but trans is nil", jsip.Name())
			return ERROR
		}

		trans = newJSIPTrans(tid, jsip, sendrecv)

		if jsip.Type == ACK {
			trans.delete()

			rid, ok := jsip.GetInt("RelatedID")
			if !ok {
				m.log.LogInfo(jsip, "ACK miss RelatedID")
				return IGNORE
			}

			ackid := m.transactionID(jsip, uint64(rid))
			ackTrans := m.transactions[ackid]
			if ackTrans == nil {
				m.log.LogInfo(jsip, "Transaction INVITE not exist")
				return IGNORE
			}

			if ackTrans.State < TRANS_FINALRESP {
				m.log.LogError(jsip,
					"Recv ACK but not receive final response")
				return IGNORE
			}

			if ackTrans.UAType == UAS && sendrecv == SEND ||
				ackTrans.UAType == UAC && sendrecv == RECV {

				m.log.LogError(jsip, "ACK direct is not same as INVITE")
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

				m.log.LogInfo(jsip, "CANCEL miss RelatedID")
				return IGNORE
			}

			cancelid := m.transactionID(jsip, uint64(rid))
			cancelTrans := m.transactions[cancelid]
			if cancelTrans == nil {
				trans.delete()

				m.log.LogInfo(jsip, "Transaction Cancelled not exist")
				return IGNORE
			}

			if cancelTrans.State >= TRANS_FINALRESP {
				trans.delete()

				m.log.LogInfo(jsip,
					"Transaction in finalize response, cannot cancel")
				return IGNORE
			}

			if cancelTrans.UAType == UAS && sendrecv == SEND ||
				cancelTrans.UAType == UAC && sendrecv == RECV {

				trans.delete()

				m.log.LogError(jsip, "CANCEL direct is not same as Request")
				return IGNORE
			}

			if sendrecv == RECV {
				// Send CANCLE 200
				if jsip.inner {
					m.sendq_t <- JSIPMsgRes(cancelTrans.req, 408)
					trans.delete()
				} else {
					m.sendq_t <- JSIPMsgRes(jsip, 200)
					m.sendq_t <- JSIPMsgRes(cancelTrans.req, 487)
				}
			}

			if cancelTrans.Type != INVITE && cancelTrans.Type != REGISTER &&
				cancelTrans.Type != OPTIONS && cancelTrans.Type != MESSAGE &&
				cancelTrans.Type != SUBSCRIBE && cancelTrans.Type != NOTIFY {

				return IGNORE
			}
		}

		if jsip.Type == BYE {
			if sendrecv == RECV {
				// Send BYE 200
				if jsip.inner {
					trans.delete()
				} else {
					m.sendq_t <- JSIPMsgRes(jsip, 200)
				}
			}
		}

		return OK
	}

	if jsip.Code == 0 {
		m.log.LogError(jsip, "process %s but trans exist", jsip.Name())
		return IGNORE
	}

	// Response
	if trans.UAType == UAS && sendrecv == RECV ||
		trans.UAType == UAC && sendrecv == SEND {

		m.log.LogError(jsip, "Response direct is same as Request direct")
		return IGNORE
	}

	jsip.Type = trans.Type

	if jsip.Code == 100 {
		if trans.State > TRANS_TRYING {
			m.log.LogError(jsip, "process 100 Trying but state is %d",
				trans.State)
			return IGNORE
		}

		trans.State = TRANS_TRYING

		return IGNORE
	}

	if jsip.Code < 200 && jsip.Code > 100 {
		if trans.State > TRANS_PR {
			m.log.LogError(jsip, "process %s but state is %d", jsip.Name(),
				trans.State)
			return IGNORE
		}

		trans.timer.Reset(m.dconfig.PRTimer)
		trans.State = TRANS_PR

		return OK
	}

	if trans.State >= TRANS_FINALRESP {
		m.log.LogError(jsip, "process %s but state is %d", jsip.Name(),
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
			m.sendq_t <- JSIPMsgAck(jsip)
		}
		// Wait for ACK
		trans.timer.Reset(m.dconfig.TransTimer)
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
func (m *JSIPStack) recvJSIPTerm(dlg string) {
	jsip := &JSIP{
		Type:       TERM,
		DialogueID: dlg,
	}

	m.jsipC <- jsip
}

func (m *JSIPStack) jsipInviteSession(session *JSIPSession, jsip *JSIP,
	sendrecv int) int {

	if jsip.Type == TERM {
		if session.State >= INVITE_ERR {
			return IGNORE
		}

		if session.State >= INVITE_200 {
			bye := JSIPMsgBye(session)
			session.cseq++
			bye.CSeq = session.cseq

			SendMsg(bye)

			return IGNORE
		}

		if session.UAType == UAC {
			cancel := JSIPMsgCancel(session.req)
			session.cseq++
			cancel.CSeq = session.cseq

			SendMsg(cancel)
		} else {
			resp := JSIPMsgRes(session.req, 487)

			SendMsg(resp)
		}

		return IGNORE
	}

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
				m.sendq_t <- JSIPMsgRes(jsip, 200)
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
				session.sessTimer = golib.NewTimer(session.expire,
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
				session.sessTimer = golib.NewTimer(session.expire,
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
		} else if jsip.Code >= 300 && jsip.Type == INVITE {
			session.State = INVITE_ACK
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

	m.log.LogError(jsip, "%s %s in %s", jsipDirect[sendrecv], jsip.Name(),
		jsipInviteState[session.State])

	return ERROR
}

func (m *JSIPStack) jsipDefaultSession(session *JSIPSession, jsip *JSIP,
	sendrecv int) int {

	if jsip.Type == CANCEL && jsip.Code > 0 {
		return IGNORE
	}

	if jsip.Type == TERM {
		if session.State >= TRANS_FINALRESP {
			return IGNORE
		}

		if session.UAType == UAC {
			cancel := JSIPMsgCancel(session.req)
			session.cseq++
			cancel.CSeq = session.cseq

			SendMsg(cancel)
		} else {
			resp := JSIPMsgRes(session.req, 487)

			SendMsg(resp)
		}

		return IGNORE
	}

	switch session.State {
	case DEFAULT_INIT:
		if jsip.Code != 0 {
			m.log.LogError(jsip,
				"Recv response %s but session state is DEFAULT_INIT",
				jsip.Name())
			return ERROR
		}

		if session.Type == CANCEL {
			m.log.LogError(jsip,
				"Session not exist when process msg %s", jsip.Name())
			session.State = DEFAULT_RESP

			resp := JSIPMsgRes(jsip, 481)
			if sendrecv == RECV {
				m.sendq_t <- resp
			} else {
				m.log.LogDebug(jsip, "Recv: %s", resp.Abstract())
				m.jsipC <- resp
			}

			return ERROR
		} else if session.Type != INVITE && session.Type != REGISTER &&
			session.Type != OPTIONS && session.Type != MESSAGE &&
			session.Type != SUBSCRIBE && session.Type != NOTIFY {

			m.log.LogError(jsip, "Session not exist when process msg %s",
				jsip.Name())
			session.State = DEFAULT_REQ

			resp := JSIPMsgRes(jsip, 481)
			if sendrecv == RECV {
				m.sendq_s <- resp
			} else {
				m.jsipC <- resp
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

			m.log.LogError(jsip,
				"Recv request %s but session state is DEFAULT_REQ", jsip.Name())

			resp := JSIPMsgRes(jsip, 400)
			if sendrecv == RECV {
				m.sendq_t <- resp
			} else {
				m.recvq_s <- resp
			}

			return ERROR
		}

		if jsip.Code >= 200 && session.req.CSeq == jsip.CSeq {
			session.State = DEFAULT_RESP
		}
	}

	return OK
}

func (m *JSIPStack) jsipSession(jsip *JSIP, sendrecv int) int {
	if !m.dconfig.SessionLayer {
		return OK
	}

	session := m.sessions[jsip.DialogueID]
	jsip.Session = session

	if session == nil {
		if jsip.Code != 0 {
			m.log.LogError(jsip, "recv response but session is nil")
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
			ret = DONE
		}
	} else {
		if session.State >= DEFAULT_RESP {
			session.delete()
			ret = DONE
		}
	}

	return ret
}

func (m *JSIPStack) recvJSIPMsg_t(jsip *JSIP) {
	// Transaction Layer
	m.log.LogDebug(jsip, "Recv[Transaction]: %s", jsip.Abstract())
	ret := m.jsipTransaction(jsip, RECV)
	if ret == ERROR {
		return
	} else if ret == IGNORE {
		return
	}

	m.recvq_s <- jsip
}

func (m *JSIPStack) recvJSIPMsg_s(jsip *JSIP) {
	// Session Layer
	m.log.LogDebug(jsip, "Recv[Session]: %s", jsip.Abstract())
	ret := m.jsipSession(jsip, RECV)
	if ret == ERROR {
		return
	} else if ret == IGNORE {
		return
	}

	m.log.LogDebug(jsip, "Recv: %s", jsip.Abstract())
	m.jsipC <- jsip

	if ret == DONE && m.dconfig.TermNotify {
		m.recvJSIPTerm(jsip.DialogueID)
	}
}

func (m *JSIPStack) sendJSIPMsg_s(jsip *JSIP) {
	// Session Layer
	m.log.LogDebug(jsip, "Send[Session]: %s", jsip.Abstract())
	ret := m.jsipSession(jsip, SEND)
	if ret == ERROR {
		return
	} else if ret == IGNORE {
		return
	}

	m.sendq_t <- jsip

	if ret == DONE && m.dconfig.TermNotify {
		m.recvJSIPTerm(jsip.DialogueID)
	}
}

func (m *JSIPStack) sendJSIPMsg_t(jsip *JSIP) {
	// Transaction Layer
	m.log.LogDebug(jsip, "Send[Transaction]: %s", jsip.Abstract())
	ret := m.jsipTransaction(jsip, SEND)
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
		var conn *golib.WSConn
		if len(jsip.Router) > 0 {
			conn = m.connect(jsip.Router[0], jsip.Userid)
		} else {
			conn = m.connect(jsip.RequestURI, jsip.Userid)
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

	jsip = m.jsipParser(jsip)

	data, _ := json.Marshal(jsip.RawMsg)
	m.log.LogDebug(jsip, "Send[Raw] %s", string(data))
	jsip.conn.Send(data)
}

// public interface

// JSIP Stack realm, such as imas.test.com
func Realm() string {
	return jstack.dconfig.Realm
}

// JSIP Stack queue to application layer
func (m *JSIPStack) JSIPChannel() <-chan *JSIP {
	return m.jsipC
}

// Recv Msg from websocket channel
func RecvMsg(conn golib.Conn, data []byte) {
	jstack.log.LogDebug(conn, "Recv[Raw]: %s", string(data))
	jsip, err := jstack.jsipUnParser(data)
	if err != nil {
		jstack.log.LogError(conn, "jsipUnParser err %s", err)
		return
	}

	jsip.conn = conn

	jstack.recvq_t <- jsip
}

// Send Msg to jsip stack
func SendMsg(jsip *JSIP) {
	jstack.log.LogDebug(jsip, "Send: %s", jsip.Abstract())
	jsip, err := jstack.jsipPrepared(jsip)
	if err != nil {
		jstack.log.LogError(jsip, "jsipPrepared err %s", err)
		return
	}

	jstack.sendq_s <- jsip
}

// Send Term msg, jsip stack will transfer term msg to JSIP end msg
// such as BYE, CACNEL, INVITE_487 according to current session state
func SendJSIPTerm(dlg string) {
	if !jstack.dconfig.SessionLayer {
		return
	}

	jsip := &JSIP{
		Type:       TERM,
		DialogueID: dlg,
	}

	SendMsg(jsip)
}

// Clone a jsip msg, using new dlg
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
		Body:       copyBody(req.Body),
		RawMsg:     copyMap(req.RawMsg),
	}

	return msg
}

// Create a response msg with code for req
func JSIPMsgRes(req *JSIP, code int) *JSIP {
	if req.Code != 0 {
		jstack.log.LogError(req, "Cannot send response for response")
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

// Create a ACK msg for resp
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

// Create a BYE msg for session
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

// Create a UPDATE msg for session
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

// Create a CACNEL msg for req
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
