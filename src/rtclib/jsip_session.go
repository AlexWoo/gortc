// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//

// JSIP Session

package rtclib

import (
	"time"

	"github.com/alexwoo/golib"
)

type jsipSessionState int

const (
	INVITE_INIT jsipSessionState = iota + 1
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

var jsipSessionStateStr = []string{
	"Unknown",
	"INVITE_INIT",
	"INVITE_18X",
	"INVITE_PRACK",
	"INVITE_UPDATE",
	"INVITE_200",
	"INVITE_ACK",
	"INVITE_REINV",
	"INVITE_RE200",
	"INVITE_ERR",
	"INVITE_END",
}

func (s jsipSessionState) String() string {
	if s < jsipSessionState(Unknown) || s > INVITE_END {
		s = jsipSessionState(Unknown)
	}

	return jsipSessionStateStr[s]
}

type jsipSession struct {
	req   *JSIP
	state jsipSessionState
	init  *jsipSessionInit
	log   *golib.Log

	inviteRecv     bool
	updateRecv     bool
	timer          *time.Timer
	sessionTimeout time.Duration
	waitUpdateResp bool
	failureCount   uint8

	msgC      chan *JSIP
	cancelled bool
}

type jsipSessionInit struct {
	sessionFailureCount uint8
	sessionTimer        time.Duration
	prTimer             time.Duration
	transTimer          time.Duration
	qsize               uint64
	msg                 chan *JSIP
	term                chan string
}

func inviteSession(m *JSIP) bool {
	switch m.Type {
	case MESSAGE:
		return false
	case SUBSCRIBE:
		return false
	case REGISTER:
		return false
	case NOTIFY:
		return false
	case OPTIONS:
		return false
	case TERM:
		return false
	default:
		return true
	}
}

func createSession(m *JSIP, init *jsipSessionInit, log *golib.Log) *jsipSession {
	s := &jsipSession{
		req:        m,
		state:      INVITE_INIT,
		init:       init,
		log:        log,
		inviteRecv: m.recv,
		msgC:       make(chan *JSIP, init.qsize),
	}

	if expire, ok := m.GetUint("Expire"); ok {
		init.sessionTimer = time.Duration(expire) * time.Second
	} else {
		// if first INVITE not set Expire, set it
		m.SetUint("Expire", uint64(s.init.sessionTimer.Seconds()))
	}

	// make sure transaction timer trigger first
	s.timer = time.NewTimer(2 * s.init.prTimer)

	go s.loop()

	return s
}

func (s *jsipSession) onMsg(m *JSIP) {
	s.msgC <- m
}

func (s *jsipSession) quit() {
	byeIn := JSIPMsgBye(s.req)
	byeIn.recv = true
	byeOut := JSIPMsgBye(s.req)

	s.init.msg <- byeIn
	s.msgC <- byeOut
}

func (s *jsipSession) loop() {
	defer func() {
		term := &JSIP{
			Type:       TERM,
			DialogueID: s.req.DialogueID,
			recv:       true,
		}
		s.init.msg <- term

		s.init.term <- s.req.DialogueID
	}()

	s.init.msg <- s.req

	for {
		select {
		case msg := <-s.msgC:
			if msg.Type == INFO {
				s.init.msg <- msg
				continue
			}

			process := s.getProcess()
			if process == nil {
				return
			}

			state, err := process(msg)
			s.state = state

			if err != nil {
				if err.Error() != "Ignore" {
					s.log.LogError(msg, "Process msg err: %s in %s", err.Error(), state.String())
				}
			} else {
				s.init.msg <- msg
			}

			if s.state == INVITE_END {
				return
			}

		case <-s.timer.C:
			if s.state < INVITE_200 {
				s.log.LogError(s.req, "Session Timeout at %s", s.state.String())

				resp := JSIPMsgRes(s.req, 408)

				if s.req.recv {
					// Send CANCEL to app layer
					cancel := JSIPMsgCancel(s.req)
					cancel.recv = true

					s.init.msg <- cancel

					// Send 408 to peer
					s.msgC <- resp
				} else {
					// Send 408 to app layer
					resp.recv = true
					s.init.msg <- resp

					return
				}

				continue
			}

			if s.state >= INVITE_ERR {
				return
			}

			// session Timeout
			if s.req.recv { // Wait for session update from peer timeout
				s.log.LogError(s.req, "Wait for session update from peer timeout")
				s.quit()
				continue
			}

			// failureCount will reset when receive UPDATE 200
			s.failureCount++
			if s.failureCount > s.init.sessionFailureCount {
				s.log.LogError(s.req, "Wait for session update 200 failed")
				s.quit()
				continue
			}

			// send UPDATE
			update := JSIPMsgUpdate(s.req)
			update.SetUint("Expire", uint64(s.init.sessionTimer.Seconds()))

			s.init.msg <- update

			s.timer.Reset(s.sessionTimeout) // Set timer for send next UPDATE
		}
	}
}
