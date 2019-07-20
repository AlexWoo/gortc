// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//

// JSIP Session Process State

package rtclib

import (
	"errors"
)

type process func(m *JSIP) (jsipSessionState, error)

func (s *jsipSession) getProcess() process {
	switch s.state {
	case INVITE_INIT:
		return s.sInit
	case INVITE_18X:
		return s.s18X
	case INVITE_PRACK:
		return s.sPrack
	case INVITE_UPDATE:
		return s.sUpdate
	case INVITE_200:
		return s.s200
	case INVITE_ACK:
		return s.sAck
	case INVITE_REINV:
		return s.sReInv
	case INVITE_RE200:
		return s.sRe200
	case INVITE_ERR:
		return s.sErr
	default:
		return nil
	}
}

func (s *jsipSession) sInit(m *JSIP) (jsipSessionState, error) {
	if m.Type == INVITE {
		typ := JSIPRespType(m.Code)
		switch typ {
		case JSIPProvisionalResp:
			return s.process1XX(m)
		case JSIPSuccessResp:
			return s.process2XX(m)
		case JSIPRedirectionResp:
			return s.processErrResp(m)
		case JSIPClientErrResp:
			return s.processErrResp(m)
		case JSIPServerErrResp:
			return s.processErrResp(m)
		case JSIPGlobalFailureResp:
			return s.processErrResp(m)
		default:
			break
		}
	} else if m.Type == CANCEL {
		return s.processCancel(m)
	}

	return s.state, errors.New("Unexpected msg")
}

func (s *jsipSession) s18X(m *JSIP) (jsipSessionState, error) {
	if m.Type == INVITE {
		typ := JSIPRespType(m.Code)
		switch typ {
		case JSIPProvisionalResp:
			return s.process1XX(m)
		case JSIPSuccessResp:
			return s.process2XX(m)
		case JSIPRedirectionResp:
			return s.processErrResp(m)
		case JSIPClientErrResp:
			return s.processErrResp(m)
		case JSIPServerErrResp:
			return s.processErrResp(m)
		case JSIPGlobalFailureResp:
			return s.processErrResp(m)
		default:
			break
		}
	} else if m.Type == CANCEL {
		return s.processCancel(m)
	} else if m.Type == PRACK {
		if m.Code == 0 {
			return s.processPrack(m)
		}
	} else if m.Type == UPDATE {
		if m.Code == 0 {
			return s.processUpdate(m)
		}
	}

	return s.state, errors.New("Unexpected msg")
}

func (s *jsipSession) sPrack(m *JSIP) (jsipSessionState, error) {
	if m.Type == CANCEL {
		return s.processCancel(m)
	} else if m.Type == PRACK {
		if m.Code != 0 {
			return s.processPrackResp(m)
		}
	}

	return s.state, errors.New("Unexpected msg")
}

func (s *jsipSession) sUpdate(m *JSIP) (jsipSessionState, error) {
	if m.Type == CANCEL {
		return s.processCancel(m)
	} else if m.Type == UPDATE {
		if m.Code != 0 {
			return s.processUpdateResp(m)
		}
	}

	return s.state, errors.New("Unexpected msg")
}

func (s *jsipSession) s200(m *JSIP) (jsipSessionState, error) {
	if m.Type == BYE {
		return s.processBye(m)
	} else if m.Type == ACK {
		return s.processAck(m)
	} else if m.Type == INVITE {
		if m.Code == 0 {
			return s.processReInvite(m)
		}
	} else if m.Type == UPDATE {
		if m.Code == 0 {
			return s.processSessionUpdate(m)
		} else {
			return s.processSessionUpdateResp(m)
		}
	}

	return s.state, errors.New("Unexpected msg")
}

func (s *jsipSession) sAck(m *JSIP) (jsipSessionState, error) {
	if m.Type == BYE {
		return s.processBye(m)
	} else if m.Type == INVITE {
		if m.Code == 0 {
			return s.processReInvite(m)
		}
	} else if m.Type == UPDATE {
		if m.Code == 0 {
			return s.processSessionUpdate(m)
		} else {
			return s.processSessionUpdateResp(m)
		}
	}

	return s.state, errors.New("Unexpected msg")
}

func (s *jsipSession) sReInv(m *JSIP) (jsipSessionState, error) {
	if m.Type == BYE {
		return s.processBye(m)
	} else if m.Type == INVITE {
		if m.Code != 0 {
			return s.processReResp(m)
		}
	} else if m.Type == UPDATE {
		if m.Code == 0 {
			return s.processSessionUpdate(m)
		} else {
			return s.processSessionUpdateResp(m)
		}
	}

	return s.state, errors.New("Unexpected msg")
}

func (s *jsipSession) sRe200(m *JSIP) (jsipSessionState, error) {
	if m.Type == BYE {
		return s.processBye(m)
	} else if m.Type == ACK {
		return s.processAck(m)
	} else if m.Type == UPDATE {
		if m.Code == 0 {
			return s.processSessionUpdate(m)
		} else {
			return s.processSessionUpdateResp(m)
		}
	}

	return s.state, errors.New("Unexpected msg")
}

func (s *jsipSession) sErr(m *JSIP) (jsipSessionState, error) {
	if m.Type == ACK {
		return s.processErrAck(m)
	}

	return s.state, errors.New("Unexpected msg")
}
