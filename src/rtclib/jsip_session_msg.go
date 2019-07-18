// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//

// JSIP Session Process Msg

package rtclib

import (
	"errors"
)

func (s *jsipSession) processSessionUpdate(m *JSIP) (jsipSessionState, error) {
	if !m.recv {
		return s.state, errors.New("Send session update")
	}

	if !s.req.recv {
		return s.state, errors.New("Session update direction")
	}

	s.timer.Reset(s.sessionTimeout)

	// Send UPDATE 200 to peer
	up200 := JSIPMsgRes(m, 200)
	s.init.msg <- up200

	return s.state, errors.New("Ignore")
}

func (s *jsipSession) processSessionUpdateResp(m *JSIP) (jsipSessionState, error) {
	if !m.recv {
		return s.state, errors.New("Send session update response")
	}

	if s.req.recv {
		return s.state, errors.New("Session update response direction")
	}

	if m.Code == 200 {
		s.failureCount = 0
	} else {
		return s.state, errors.New("Session update unexpected response")
	}

	return s.state, errors.New("Ignore")
}

func (s *jsipSession) process1XX(m *JSIP) (jsipSessionState, error) {
	if m.recv == s.req.recv {
		return s.state, errors.New("Response direction")
	}

	if s.cancelled {
		return s.state, errors.New("Ignore")
	}

	if m.Code == 100 {
		return s.state, errors.New("Ignore")
	}

	return INVITE_18X, nil
}

func (s *jsipSession) process2XX(m *JSIP) (jsipSessionState, error) {
	if m.recv == s.req.recv {
		return s.state, errors.New("Response direction")
	}

	if m.Code != 200 {
		s.quit()
		return INVITE_END, errors.New("Unexpected response")
	}

	if s.cancelled {
		return s.state, errors.New("Ignore")
	}

	if s.req.recv { // Wait for UPDATE from peer
		s.sessionTimeout = s.init.sessionTimer
	} else { // Send UPDATE to peer
		s.sessionTimeout = s.init.sessionTimer / 3
	}

	s.timer.Reset(s.sessionTimeout)

	return INVITE_200, nil
}

func (s *jsipSession) processErrResp(m *JSIP) (jsipSessionState, error) {
	if m.recv == s.req.recv {
		return s.state, errors.New("Response direction")
	}

	if m.recv {
		s.init.msg <- JSIPMsgAck(m)
		return INVITE_END, nil
	} else {
		s.timer.Reset(s.init.transTimer)
		return INVITE_ERR, nil
	}
}

func (s *jsipSession) processErrAck(m *JSIP) (jsipSessionState, error) {
	if m.recv != s.req.recv {
		return s.state, errors.New("ACK direction")
	}

	return INVITE_END, errors.New("Ignore")
}

func (s *jsipSession) processCancel(m *JSIP) (jsipSessionState, error) {
	if m.recv != s.req.recv {
		return s.state, errors.New("CANCEL direction")
	}

	if s.cancelled {
		return s.state, errors.New("Ignore")
	}

	if m.recv {
		s.init.msg <- JSIPMsgRes(m, 200)
		s.init.msg <- JSIPMsgRes(s.req, 487)

		return INVITE_ERR, nil
	} else {
		s.cancelled = true
		return s.state, nil
	}
}

func (s *jsipSession) processPrack(m *JSIP) (jsipSessionState, error) {
	if m.recv != s.req.recv {
		return s.state, errors.New("PRACK direction")
	}

	if s.cancelled {
		return s.state, errors.New("Ignore")
	}

	return INVITE_PRACK, nil
}

func (s *jsipSession) processPrackResp(m *JSIP) (jsipSessionState, error) {
	if m.recv == s.req.recv {
		return s.state, errors.New("PRACK response direction")
	}

	if s.cancelled {
		return s.state, errors.New("Ignore")
	}

	typ := JSIPRespType(m.Code)
	switch typ {
	case JSIPProvisionalResp:
		return s.state, errors.New("PRACK provisional response")
	case JSIPSuccessResp:
		return INVITE_18X, nil
	default: // Error Response will rollback
		return INVITE_18X, errors.New("PRACK error response")
	}
}

func (s *jsipSession) processUpdate(m *JSIP) (jsipSessionState, error) {
	if s.cancelled {
		return s.state, errors.New("Ignore")
	}

	s.updateRecv = m.recv

	return INVITE_UPDATE, nil
}

func (s *jsipSession) processUpdateResp(m *JSIP) (jsipSessionState, error) {
	if m.recv == s.updateRecv {
		return s.state, errors.New("UPDATE response direction")
	}

	if s.cancelled {
		return s.state, errors.New("Ignore")
	}

	typ := JSIPRespType(m.Code)
	switch typ {
	case JSIPProvisionalResp:
		return s.state, errors.New("UPDATE provisional response")
	case JSIPSuccessResp:
		return INVITE_18X, nil
	default: // Error Response will rollback
		return INVITE_18X, errors.New("UPDATE error response")
	}
}

func (s *jsipSession) processAck(m *JSIP) (jsipSessionState, error) {
	if m.recv != s.inviteRecv {
		return s.state, errors.New("ACK direction")
	}

	return INVITE_ACK, nil
}

func (s *jsipSession) processBye(m *JSIP) (jsipSessionState, error) {
	if m.recv {
		s.init.msg <- JSIPMsgRes(m, 200)
	}

	return INVITE_END, nil
}

func (s *jsipSession) processReInvite(m *JSIP) (jsipSessionState, error) {
	s.inviteRecv = m.recv

	return INVITE_REINV, nil
}

func (s *jsipSession) processReResp(m *JSIP) (jsipSessionState, error) {
	if m.recv == s.inviteRecv {
		return s.state, errors.New("Re-INVITE response direction")
	}

	typ := JSIPRespType(m.Code)
	switch typ {
	case JSIPProvisionalResp:
		return s.state, errors.New("Re-INVITE provisional response")
	case JSIPSuccessResp:
		return INVITE_RE200, nil
	default: // Error Response will rollback
		if m.recv {
			s.init.msg <- JSIPMsgAck(m)
			return INVITE_ACK, nil
		}

		return INVITE_RE200, nil
	}
}
