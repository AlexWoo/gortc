// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//

// JSIP Type

package rtclib

// JSIP Type
type JSIPType int

const (
	// as INVITE in SIP
	INVITE JSIPType = iota + 1

	// as ACK in SIP
	ACK

	// as BYE in SIP
	BYE

	// as CANCEL in SIP
	CANCEL

	// as REGISTER in SIP
	REGISTER

	// as OPTIONS in SIP
	OPTIONS

	// as INFO in SIP
	INFO

	// as UPDATE in SIP
	UPDATE

	// as PRACK in SIP
	PRACK

	// as SUBSCRIBE in SIP
	SUBSCRIBE

	// as MESSAGE in SIP
	MESSAGE

	// as NOTIFY in SIP
	NOTIFY

	// inner message for indicates a JSIP Session has been terminated
	TERM
)

var jsipTypeStr = []string{
	"UNKNOWN",
	"INVITE",
	"ACK",
	"BYE",
	"CANCEL",
	"REGISTER",
	"OPTIONS",
	"INFO",
	"UPDATE",
	"PRACK",
	"SUBSCRIBE",
	"MESSAGE",
	"NOTIFY",
	"TERM",
}

var jsipType = map[string]JSIPType{
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

// Generate JSIPType by string
func NewJSIPType(raw string) JSIPType {
	return jsipType[raw]
}

// Return string for JSIPType
func (t JSIPType) String() string {
	if t < JSIPType(Unknown) || t > TERM {
		return "UNKNOWN"
	}

	return jsipTypeStr[t]
}
