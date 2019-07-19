// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//

// JSIP Stack

package rtclib

import (
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"github.com/alexwoo/golib"
)

var (
	once sync.Once

	jstack *JSIPStack
)

type jsipDConfig struct {
	Realm        string        `default:"gortc.com"`
	Location     string        `default:"/rtc"`
	Qsize        uint64        `default:"1024"`
	ConnTimeout  time.Duration `default:"3s"`
	Retry        int64         `default:"10"`
	TransTimer   time.Duration `default:"5s"`
	PRTimer      time.Duration `default:"60s"`
	SessionTimer time.Duration `default:"600s"`
}

type JSIPStack struct {
	log      *golib.Log
	logLevel int

	config *jsipDConfig

	recvq chan *JSIP

	sendq chan *JSIP

	transq   chan *JSIP
	tranTerm chan string

	sessq    chan *JSIP
	sessTerm chan string

	handler func(*JSIP)

	connLock     sync.Mutex
	conns        map[string]golib.Conn
	sessLock     sync.Mutex
	sessions     map[string]*jsipSession
	transLock    sync.Mutex
	transactions map[string]*jsipTransaction
}

func JStackInstance() *JSIPStack {
	once.Do(func() {
		jstack = &JSIPStack{
			conns:        map[string]golib.Conn{},
			transactions: map[string]*jsipTransaction{},
			sessions:     map[string]*jsipSession{},
		}

		if err := jstack.loadDConfig(); err != nil {
			jstack = nil
			return
		}

		jstack.recvq = make(chan *JSIP, jstack.config.Qsize)

		jstack.sendq = make(chan *JSIP, jstack.config.Qsize)

		jstack.transq = make(chan *JSIP, jstack.config.Qsize)
		jstack.tranTerm = make(chan string, jstack.config.Qsize)

		jstack.sessq = make(chan *JSIP, jstack.config.Qsize)
		jstack.sessTerm = make(chan string, jstack.config.Qsize)

		go jstack.loop()
	})

	return jstack
}

func (s *JSIPStack) loadDConfig() error {
	confPath := FullPath("conf/gortc.ini")

	config := &jsipDConfig{}
	err := golib.ConfigFile(confPath, "JSIPStack", config)
	if err != nil {
		return fmt.Errorf("Parse dconfig %s Failed, %s", confPath, err)
	}
	s.config = config

	return nil
}

func (s *JSIPStack) SetLog(log *golib.Log, logLevel int) {
	s.log = log
	s.logLevel = logLevel
}

func (s *JSIPStack) SetHandler(h func(*JSIP)) {
	s.handler = h
}

func (s *JSIPStack) State() string {
	output := ""

	s.connLock.Lock()
	output += "!!!!! connections: " + strconv.Itoa(len(s.conns)) + "\n"
	for dlg := range s.conns {
		output += fmt.Sprintf("\t%s\n", dlg)
	}
	s.connLock.Unlock()

	s.sessLock.Lock()
	output += "!!!!! sessions: " + strconv.Itoa(len(s.sessions)) + "\n"
	for dlg := range s.sessions {
		output += fmt.Sprintf("\t%s\n", dlg)
	}
	s.sessLock.Unlock()

	s.transLock.Lock()
	output += "!!!!! transactions: " + strconv.Itoa(len(s.transactions)) + "\n"
	for tid := range s.transactions {
		output += fmt.Sprintf("\t%s\n", tid)
	}
	s.transLock.Unlock()

	return output
}

func (s *JSIPStack) processTransaction(msg *JSIP) {
	tid := transactionID(msg.DialogueID, msg.CSeq)
	s.transLock.Lock()
	trans := s.transactions[tid]
	s.transLock.Unlock()

	if msg.Code == 0 {
		if trans != nil {
			s.log.LogError(msg, "process request but transaction exists")
			return
		}

		init := &jsipTransInit{
			transTimer: s.config.TransTimer,
			prTimer:    s.config.PRTimer,
			qsize:      s.config.Qsize,
			msg:        s.transq,
			term:       s.tranTerm,
		}

		if msg.conn != nil {
			if msg.Type == INVITE || !inviteSession(msg) {
				s.connLock.Lock()
				s.conns[msg.DialogueID] = msg.conn
				s.connLock.Unlock()
			}
		}

		trans = createTransaction(msg, init, s.log)

		s.transLock.Lock()
		s.transactions[tid] = trans
		s.transLock.Unlock()
	} else {
		if trans == nil {
			s.log.LogError(msg, "process response but transaction not exists")
			return
		}

		if msg.Type == JSIPType(Unknown) {
			msg.Type = trans.req.Type
		}

		trans.onMsg(msg)
	}
}

// pre process msg from application layer
func (s *JSIPStack) preProcess(msg *JSIP) error {
	msg.recv = false

	if msg.rawMsg == nil {
		msg.rawMsg = make(map[string]interface{})
	}

	if msg.DialogueID == "" {
		return errors.New("msg no DialogueID")
	}

	s.connLock.Lock()
	msg.conn = s.conns[msg.DialogueID]
	s.connLock.Unlock()

	typ := JSIPRespType(msg.Code)
	if typ == JSIPResponseType(Unknown) {
		return errors.New("Unknown code")
	}

	if typ == JSIPReq {
		if msg.RequestURI == "" {
			return errors.New("Request no RequestURI")
		}

		if msg.From == "" {
			return errors.New("Request no From")
		}

		if msg.To == "" {
			return errors.New("Request no To")
		}

		if msg.CSeq == 0 {
			msg.CSeq = uint64(rand.Uint32())
		}

		return nil
	} else {
		if msg.CSeq == 0 {
			return errors.New("Response no CSeq")
		}

		tid := transactionID(msg.DialogueID, msg.CSeq)
		s.transLock.Lock()
		trans := s.transactions[tid]
		s.transLock.Unlock()
		if trans == nil {
			return errors.New("Response no transaction")
		}

		req := trans.req

		if msg.Type == JSIPType(Unknown) {
			msg.Type = req.Type
		}

		if msg.From == "" {
			msg.From = req.From
		}

		if msg.To == "" {
			msg.To = req.To
		}

		return nil
	}
}

func (s *JSIPStack) processSession(msg *JSIP) {
	s.sessLock.Lock()
	sess := s.sessions[msg.DialogueID]
	s.sessLock.Unlock()
	if sess == nil {
		if msg.Type == INVITE && msg.Code == 0 {
			init := &jsipSessionInit{
				sessionFailureCount: 3,
				sessionTimer:        s.config.SessionTimer,
				prTimer:             s.config.PRTimer,
				transTimer:          s.config.TransTimer,
				qsize:               s.config.Qsize,
				msg:                 s.sessq,
				term:                s.sessTerm,
			}

			sess = createSession(msg, init, s.log)
			s.sessLock.Lock()
			s.sessions[msg.DialogueID] = sess
			s.sessLock.Unlock()

			return
		} else {
			s.log.LogError(msg, "Recv msg but session[%s] does not exists", msg.DialogueID)

			if msg.Code != 0 { // Response no session
				return
			}

			if msg.Type == ACK || msg.Type == TERM { // ACK or TERM
				return
			}

			if (msg.Type == BYE || msg.Type == CANCEL) && !msg.recv { // send BYE
				return
			}

			s.log.LogInfo(msg, "Send 481 for msg")

			res := JSIPMsgRes(msg, 481)
			if !msg.recv {
				res.recv = true
			}
			s.sessq <- res

			return
		}
	}

	msg.conn = sess.req.conn

	sess.onMsg(msg)
}

func (s *JSIPStack) connect(msg *JSIP) golib.Conn {
	s.connLock.Lock()
	conn := s.conns[msg.DialogueID]
	s.connLock.Unlock()
	if conn != nil {
		return conn
	}

	if msg.Code != 0 {
		s.log.LogError(msg, "Response cannot find connection to send")
		return nil
	}

	uri := msg.RequestURI
	if len(msg.Router) > 0 {
		uri = msg.Router[0]
	}

	jsipUri, err := NewJSIPUri(uri)
	if err != nil {
		s.log.LogError(msg, "JSIPUri unmarshal err: %s", err.Error())
		return nil
	}

	userid := msg.Userid
	if msg.Userid == "" {
		userid = s.config.Realm
	}

	url := "ws://" + jsipUri.HostportString() + s.config.Location + "?userid=" + userid
	timeout := s.config.ConnTimeout
	retry := int(s.config.Retry)
	qsize := s.config.Qsize

	return golib.NewWSClient(jsipUri.UserHostString(), url, timeout, retry, qsize, RecvMsg, s.log, s.logLevel)
}

func (s *JSIPStack) send(msg *JSIP) {
	if msg.conn == nil {
		if msg.conn = s.connect(msg); msg.conn == nil {
			return
		}
	}

	data, err := msg.Marshal()
	if err != nil {
		s.log.LogError(msg, "Marshal JSIP err: %s", err.Error())
	}

	msg.conn.Send(data)
}

func (s *JSIPStack) loop() {
	for {
		select {
		case msg := <-s.recvq:
			// TODO replace DialogueID
			s.processTransaction(msg)

		case msg := <-s.sendq:
			if err := s.preProcess(msg); err != nil {
				s.log.LogError(msg, "Pre process msg from applicaion layer error: %s", err.Error())
				continue
			}

			if inviteSession(msg) {
				s.processSession(msg)
			} else {
				s.processTransaction(msg)
			}

		case msg := <-s.transq:
			if msg.recv {
				if inviteSession(msg) {
					s.processSession(msg)
				} else {
					s.handler(msg)
				}
			} else {
				s.send(msg)
			}

			if msg.Type == TERM {
				s.connLock.Lock()
				delete(s.conns, msg.DialogueID)
				s.connLock.Unlock()
			}

		case tid := <-s.tranTerm:
			s.transLock.Lock()
			delete(s.transactions, tid)
			s.transLock.Unlock()

		case msg := <-s.sessq:
			if msg.recv {
				s.handler(msg)
			} else {
				s.processTransaction(msg)
			}

			if msg.Type == TERM {
				s.connLock.Lock()
				delete(s.conns, msg.DialogueID)
				s.connLock.Unlock()
			}

		case sid := <-s.sessTerm:
			s.sessLock.Lock()
			delete(s.sessions, sid)
			s.sessLock.Unlock()
		}
	}
}

func Realm() string {
	return jstack.config.Realm
}

func RecvMsg(conn golib.Conn, data []byte) {
	m := &JSIP{
		conn: conn,
		recv: true,
	}

	if err := m.Unmarshal(data); err != nil {
		jstack.log.LogError(conn, "Unmarshal JSIP msg error: %s", err.Error())
		return
	}

	jstack.recvq <- m
}

func SendMsg(m *JSIP) {
	if m == nil {
		jstack.log.LogError(m, "SendMsg, m is nil")
		return
	}

	jstack.sendq <- m
}
