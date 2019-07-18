// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//

// JSIP Message

package rtclib

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/alexwoo/golib"
	"github.com/tidwall/gjson"
)

type JSIP struct {
	Type       JSIPType
	Code       int
	RequestURI string
	From       string
	To         string
	CSeq       uint64
	DialogueID string
	Router     []string
	Body       interface{}

	Userid string
	Term   bool

	conn   golib.Conn
	rawMsg map[string]interface{}
	recv   bool
}

// for log ctx

// Log ctx Prefix
func (m *JSIP) Prefix() string {
	return "[jstack]"
}

// Log ctx Suffix
func (m *JSIP) Suffix() string {
	if m == nil {
		return "[jsip] nil"
	}

	suf := "[jsip] " + m.String()

	if m.conn != nil {
		suf += " [conn] " + m.conn.Suffix()
	}

	return suf
}

// Log ctx LogLevel
func (m *JSIP) LogLevel() int {
	if m != nil && m.conn != nil {
		return m.conn.LogLevel()
	}

	if jstack != nil {
		return jstack.logLevel
	}

	return golib.LOGINFO
}

// Marshal JSIP Message to []byte
func (m *JSIP) Marshal() ([]byte, error) {
	typ := JSIPRespType(m.Code)
	if typ == JSIPResponseType(Unknown) {
		return nil, errors.New("Invalid Code")
	} else if typ == JSIPReq {
		m.rawMsg["Type"] = m.Type.String()
		if m.rawMsg["Type"] == "UNKNOWN" {
			return nil, errors.New("Unknow Type")
		}

		if m.RequestURI == "" {
			return nil, errors.New("No RequestURI")
		}
		m.rawMsg["Request-URI"] = m.RequestURI

		router := ""
		if len(m.Router) > 0 {
			router = m.Router[0]
			for i := 1; i < len(m.Router); i++ {
				router += ", " + m.Router[i]
			}
		}

		if router == "" {
			delete(m.rawMsg, "Router")
		} else {
			m.rawMsg["Router"] = router
		}
	} else {
		m.rawMsg["Type"] = "RESPONSE"
		m.rawMsg["Code"] = m.Code
		m.rawMsg["Desc"] = JSIPRespDesc(m.Code)
	}

	if m.From == "" {
		return nil, errors.New("No From")
	}
	m.rawMsg["From"] = m.From

	if m.To == "" {
		return nil, errors.New("No To")
	}
	m.rawMsg["To"] = m.To

	if m.CSeq == 0 {
		return nil, errors.New("No CSeq")
	}
	m.rawMsg["CSeq"] = m.CSeq

	if m.DialogueID == "" {
		return nil, errors.New("No DialogueID")
	}
	m.rawMsg["DialogueID"] = m.DialogueID

	if m.Body != nil {
		m.rawMsg["Body"] = m.Body
	}

	return json.Marshal(m.rawMsg)
}

// Unmarshal JSIP Message from []byte
func (m *JSIP) Unmarshal(raw []byte) error {
	rawMsg, ok := gjson.ParseBytes(raw).Value().(map[string]interface{})
	if !ok {
		return errors.New("raw is not json object")
	}

	m.rawMsg = rawMsg

	typ, ok := rawMsg["Type"].(string)
	if !ok {
		return errors.New("Type error")
	}

	if typ == "RESPONSE" {
		if m.Code, ok = getJsonInt(rawMsg, "Code"); !ok {
			return errors.New("Code error")
		}

		if JSIPRespType(m.Code) < JSIPProvisionalResp {
			return errors.New("Invalid Code")
		}
	} else {
		m.Type = NewJSIPType(typ)
		if m.Type == JSIPType(Unknown) {
			return errors.New("Unknown Type")
		}

		if m.RequestURI, ok = rawMsg["Request-URI"].(string); !ok {
			return errors.New("Request-URI error")
		}

		if routers, ok := rawMsg["Router"].(string); ok {
			m.Router = strings.Split(routers, ",")
			for i := 0; i < len(m.Router); i++ {
				m.Router[i] = strings.TrimSpace(m.Router[i])
			}
		}
	}

	if m.From, ok = rawMsg["From"].(string); !ok {
		return errors.New("From error")
	}

	if m.To, ok = rawMsg["To"].(string); !ok {
		return errors.New("To error")
	}

	if m.CSeq, ok = getJsonUint64(rawMsg, "CSeq"); !ok {
		return errors.New("CSeq error")
	}

	if m.DialogueID, ok = rawMsg["DialogueID"].(string); !ok {
		return errors.New("DialogueID error")
	}

	body := gjson.GetBytes(raw, "Body")
	if body.Exists() {
		m.Body = body.Value()
	}

	return nil
}

// Return JSIP Message Name, Request as INVITE, Response as INVITE_200
func (m *JSIP) Name() string {
	if m.Code != 0 {
		return fmt.Sprintf("%s_%d", m.Type.String(), m.Code)
	} else {
		return m.Type.String()
	}
}

// Return JSIP Message Main para as string
func (m *JSIP) String() string {
	output := m.Name()

	if m.Type == TERM {
		output += " DialogueID: " + m.DialogueID
		return output
	}

	if m.Code == 0 {
		output += " RequestURI: " + m.RequestURI
	}

	output += " From: " + m.From
	output += " To: " + m.To
	output += " CSeq: " + strconv.FormatUint(m.CSeq, 10)
	output += " DialogueID: " + m.DialogueID

	if len(m.Router) > 0 {
		output += " Router: " + m.Router[0]
		for i := 1; i < len(m.Router); i++ {
			output += ", " + m.Router[i]
		}
	}

	if relatedid, ok := m.GetUint("RelatedID"); ok {
		output += " RelatedID: " + strconv.FormatUint(relatedid, 10)
	}

	return output
}

// Get a header with int value from JSIP Message
func (m *JSIP) GetInt(header string) (int64, bool) {
	return getJsonInt64(m.rawMsg, header)
}

// Set a header with int value to JSIP Message
func (m *JSIP) SetInt(header string, value int64) {
	m.rawMsg[header] = float64(value)
}

// Get a header with uint value from JSIP Message
func (m *JSIP) GetUint(header string) (uint64, bool) {
	return getJsonUint64(m.rawMsg, header)
}

// Set a header with uint value to JSIP Message
func (m *JSIP) SetUint(header string, value uint64) {
	m.rawMsg[header] = float64(value)
}

// Get a header with string value from JSIP Message
func (m *JSIP) GetString(header string) (string, bool) {
	v, ok := m.rawMsg[header].(string)

	return v, ok
}

// Delete a header from JSIP Message
func (m *JSIP) DelHeader(header string) {
	delete(m.rawMsg, header)
}

// Set a header with string value to JSIP Message
func (m *JSIP) SetString(header string, value string) {
	m.rawMsg[header] = value
}

// Clone a jsip msg, using new dlg
func JSIPMsgClone(m *JSIP, dlg string) *JSIP {
	msg := &JSIP{
		Type:       m.Type,
		Code:       m.Code,
		RequestURI: m.RequestURI,
		From:       m.From,
		To:         m.To,
		DialogueID: dlg,
		Router:     m.Router,
		Body:       copyBody(m.Body),

		rawMsg: copyMap(m.rawMsg),
	}

	if msg.Code == 0 {
		msg.CSeq = uint64(rand.Uint32())
	}

	return msg
}

// Create a JSIP Request
func JSIPMsgReq(typ JSIPType, requestURI string, from string, to string, dlg string) *JSIP {
	msg := &JSIP{
		Type:       typ,
		RequestURI: requestURI,
		From:       from,
		To:         to,
		CSeq:       uint64(rand.Uint32()),
		DialogueID: dlg,

		rawMsg: make(map[string]interface{}),
	}

	return msg
}

// Create a JSIP Response with code for req
func JSIPMsgRes(m *JSIP, code int) *JSIP {
	rt := JSIPRespType(code)
	if rt == JSIPReq || rt == JSIPResponseType(Unknown) {
		fmt.Println("Invalid JSIP Response code")
		return nil
	}

	if JSIPRespType(m.Code) != JSIPReq {
		fmt.Println("Cannot create JSIP Response from non Request")
		return nil
	}

	msg := &JSIP{
		Type:       m.Type,
		Code:       code,
		From:       m.From,
		To:         m.To,
		CSeq:       m.CSeq,
		DialogueID: m.DialogueID,

		conn:   m.conn,
		rawMsg: make(map[string]interface{}),
	}

	return msg
}

// Create a ACK msg for resp
func JSIPMsgAck(m *JSIP) *JSIP {
	if m.Type != INVITE || JSIPRespType(m.Code) < JSIPSuccessResp {
		fmt.Println("Cannot create ACK for non INVITE final response")
		return nil
	}

	msg := JSIPMsgReq(ACK, m.To, m.From, m.To, m.DialogueID)
	msg.SetUint("RelatedID", m.CSeq)
	msg.conn = m.conn

	return msg
}

// Create a CACNEL msg for req
func JSIPMsgCancel(m *JSIP) *JSIP {
	if m.Type != INVITE || m.Code != 0 {
		fmt.Println("Cannot create CANCEL for non INVITE Request")
		return nil
	}

	msg := JSIPMsgReq(CANCEL, m.RequestURI, m.From, m.To, m.DialogueID)
	msg.SetUint("RelatedID", m.CSeq)
	msg.conn = m.conn

	return msg
}

// Create a BYE msg for session
func JSIPMsgBye(req *JSIP) *JSIP {
	if req.Type != INVITE || req.Code != 0 {
		fmt.Println("Cannot create BYE for non INVITE Request")
		return nil
	}

	bye := &JSIP{
		Type:       BYE,
		RequestURI: req.RequestURI,
		From:       req.From,
		To:         req.To,
		CSeq:       uint64(rand.Uint32()),
		DialogueID: req.DialogueID,

		conn:   req.conn,
		rawMsg: make(map[string]interface{}),
	}

	return bye
}

// Create a UPDATE msg for session
func JSIPMsgUpdate(req *JSIP) *JSIP {
	if req.Type != INVITE || req.Code != 0 {
		fmt.Println("Cannot create BYE for non INVITE Request")
		return nil
	}

	update := &JSIP{
		Type:       UPDATE,
		RequestURI: req.RequestURI,
		From:       req.From,
		To:         req.To,
		CSeq:       uint64(rand.Uint32()),
		DialogueID: req.DialogueID,

		conn:   req.conn,
		rawMsg: make(map[string]interface{}),
	}

	return update
}
