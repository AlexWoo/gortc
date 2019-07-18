// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//

// JSIP Default Transaction

package rtclib

import (
	"errors"
	"strconv"
	"time"

	"github.com/alexwoo/golib"
)

type jsipTransState int

const (
	TRANS_INIT jsipTransState = iota + 1
	TRANS_PROVISIONALRESP
	TRANS_SUCCESSESP
	TRANS_ERRRESP
)

var jsipTransStateStr = []string{
	"Unknown",
	"Init",
	"Provisional Response",
	"Success Response",
	"Error Response",
}

func (s jsipTransState) String() string {
	if s < jsipTransState(Unknown) || s > TRANS_ERRRESP {
		s = jsipTransState(Unknown)
	}

	return jsipTransStateStr[s]
}

type jsipTransInit struct {
	transTimer time.Duration
	prTimer    time.Duration
	qsize      uint64
	msg        chan *JSIP
	term       chan string
}

type jsipTransaction struct {
	req   *JSIP
	state jsipTransState
	init  *jsipTransInit
	timer *golib.Timer
	log   *golib.Log
}

func transactionID(dlg string, seq uint64) string {
	return dlg + ":" + strconv.FormatUint(seq, 10)
}

func createTransaction(m *JSIP, init *jsipTransInit, log *golib.Log) *jsipTransaction {
	if m.Code != 0 {
		return nil
	}

	t := &jsipTransaction{
		req:   m,
		state: TRANS_INIT,
		init:  init,
		log:   log,
	}

	t.init.msg <- t.req

	if m.Type == ACK {
		tid := transactionID(t.req.DialogueID, t.req.CSeq)
		t.init.term <- tid
	} else {
		t.timer = golib.NewTimer(t.init.transTimer, t.timerHandle, nil)
	}

	return t
}

// paras:
//    req: transaction request
//    m: msg received
//    s: transaction current state
// return:
//    state: transaction new state
//    err: nil, continue process, otherwise, ignore msg received
func (t *jsipTransaction) transProcess(m *JSIP) (jsipTransState, error) {
	typ := JSIPRespType(m.Code)
	if typ <= JSIPReq {
		return t.state, errors.New("Unexpected msg")
	}

	if m.Type != t.req.Type {
		return t.state, errors.New("Unexpected response")
	}

	if m.CSeq != t.req.CSeq {
		return t.state, errors.New("Unexpected CSeq")
	}

	if t.req.recv == m.recv {
		return t.state, errors.New("Unexpected direction")
	}

	if typ == JSIPProvisionalResp {
		if t.state > TRANS_PROVISIONALRESP {
			return t.state, errors.New("Unexpected provisional response")
		}

		return TRANS_PROVISIONALRESP, nil
	}

	// Final Response

	if t.state >= TRANS_SUCCESSESP {
		return t.state, errors.New("Unexpected final response")
	}

	if typ == JSIPSuccessResp {
		return TRANS_SUCCESSESP, nil
	}

	return TRANS_ERRRESP, nil
}

func (t *jsipTransaction) onMsg(m *JSIP) {
	state, err := t.transProcess(m)
	if err != nil {
		t.log.LogError(m, "%s Transaction process error in %s: %s", t.req.Type.String(), t.state.String(), err.Error())
		return
	}

	if state == TRANS_PROVISIONALRESP {
		t.timer.Reset(t.init.prTimer)
	}

	t.state = state

	if (t.req.Type != BYE && t.req.Type != CANCEL) || !m.recv { // Recv BYE response will not send to session layer
		t.init.msg <- m
	}

	if t.state >= TRANS_SUCCESSESP {
		t.quit()
	}
}

func (t *jsipTransaction) quit() {
	t.timer.Stop()

	if !inviteSession(t.req) {
		term := &JSIP{
			Type:       TERM,
			DialogueID: t.req.DialogueID,
			recv:       true,
		}
		t.init.msg <- term
	}

	tid := transactionID(t.req.DialogueID, t.req.CSeq)
	t.init.term <- tid
}

func (t *jsipTransaction) timerHandle(d interface{}) {
	t.log.LogError(t.req, "%s Transaction timeout", t.req.Type.String())

	if (t.req.Type != BYE && t.req.Type != CANCEL) || t.req.recv { // Recv BYE response will not send to session layer
		resp := JSIPMsgRes(t.req, 408)
		resp.recv = !t.req.recv

		t.init.msg <- resp
	}

	t.quit()
}
