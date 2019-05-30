// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//
// go rtc script manager

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/gorilla/websocket"
	"github.com/tidwall/gjson"
)

const (
	SEND = iota
	RECV
	IF
	END
)

const (
	ERROR = -1
	OK    = 0
)

func deepcopy(value map[string]interface{}) map[string]interface{} {
	ret := make(map[string]interface{})
	for k, v := range value {
		ret[k] = v
	}

	return ret
}

type stage struct {
	typ     int
	timeout time.Duration
	tag     string
	option  bool
	skip    string
	value   map[string]interface{}
	vars    map[string]string
}

type rtcTest struct {
	stages   []*stage
	failed   uint32
	successd uint32
	tags     map[string]int
}

func NewRTCTest(file string) *rtcTest {
	rtctest := &rtcTest{
		tags: make(map[string]int),
	}
	rtctest.compile(file)

	return rtctest
}

func (s *stage) getValue(st map[string]interface{}) {
	value, ok := st["value"].(map[string]interface{})
	if !ok {
		log.Fatalln("value", st["value"], "not map")
	}

	s.value = value
}

func (s *stage) getTimeout(st map[string]interface{}) {
	timeout, ok := st["timeout"].(string)
	if ok {
		t, err := time.ParseDuration(timeout)
		if err != nil {
			log.Fatalln("timeout not time.Duration")
		}

		s.timeout = t
	} else {
		if st["timeout"] != nil {
			log.Fatalln("timeout", st["timeout"], "not string")
		}
	}

}

func (s *stage) getTag(st map[string]interface{}) {
	tag, ok := st["tag"].(string)
	if ok {
		s.tag = tag
	} else {
		if st["tag"] != nil {
			log.Fatalln("tag", st["tag"], "not string")
		}
	}
}

func (s *stage) getOption(st map[string]interface{}) {
	opt, ok := st["option"].(bool)
	if ok {
		s.option = opt
	} else {
		if st["option"] != nil {
			log.Fatalln("option", st["option"], "not bool")
		}
	}
}

func (s *stage) getSkip(st map[string]interface{}) {
	skip, ok := st["goto"].(string)
	if ok {
		s.skip = skip
	} else {
		if st["goto"] != nil {
			log.Fatalln("goto", st["goto"], "not string")
		}
	}
}

func (s *stage) getVars(st map[string]interface{}) {
	vars, ok := st["vars"].(map[string]interface{})
	if ok {
		reg := regexp.MustCompile("^[$]{1}[a-zA-Z_]+[\\w]*$")
		for k, tmp := range vars {
			v, ok := tmp.(string)
			if !ok {
				log.Fatalln("vars", k, "value:", tmp, "not string")
			}

			switch s.typ {
			case SEND:
				// string: variable
				if !reg.MatchString(v) {
					log.Fatalln(v, "not a variable")
				}

				s.vars[v] = k
			case RECV:
				// variable: string
				if !reg.MatchString(k) {
					log.Fatalln(k, "not a variable")
				}

				s.vars[k] = v
			case IF:
				// variable: reg pattern
				if !reg.MatchString(k) {
					log.Fatalln(k, "not a variable")
				}

				s.vars[k] = v
			}
		}
	} else {
		if st["vars"] != nil {
			log.Fatalln("vars", st["vars"], "not map[string]interface{}")
		}
	}
}

func (t *rtcTest) compile(file string) {
	f, err := os.Open(file)
	if err != nil {
		log.Fatalln("open file", file, "error,", err)
	}
	defer f.Close()

	json, _ := ioutil.ReadAll(f)
	if !gjson.ValidBytes(json) {
		log.Fatalln("Invalid json", string(json))
	}

	s, ok := gjson.ParseBytes(json).Value().([]interface{})
	if !ok {
		log.Fatalln(file, "not a json array")
	}

	for i := 0; i < len(s); i++ {
		stage := &stage{
			vars: make(map[string]string),
		}

		st, ok := s[i].(map[string]interface{})
		if !ok {
			log.Fatalln(st, "not a json object")
		}

		_, ok = st["type"].(string)
		if !ok {
			log.Fatalln("type", st["type"], "not string")
		}

		switch st["type"].(string) {
		case "send":
			stage.typ = SEND
			stage.getTimeout(st)
			stage.getTag(st)
			stage.getValue(st)
			stage.getVars(st)

		case "recv":
			stage.typ = RECV
			stage.getTimeout(st)
			stage.getTag(st)
			stage.getOption(st)
			stage.getValue(st)
			stage.getVars(st)

		case "if":
			stage.typ = IF
			stage.getVars(st)
			stage.getSkip(st)

		case "end":
			stage.typ = END

		default:
			log.Fatalln("unknown type", st["type"])
		}

		if stage.tag != "" {
			t.tags[stage.tag] = len(t.stages)
		}
		t.stages = append(t.stages, stage)
	}
}

func (t *rtcTest) processSend(s *stage, conn *websocket.Conn,
	vars map[string]interface{}) int {

	value := deepcopy(s.value)

	if s.timeout > 0 {
		time.Sleep(s.timeout)
	}

	for k, v := range s.vars {
		variable, ok := vars[k]
		if !ok {
			continue
		}

		value[v] = variable
	}

	data, _ := json.Marshal(value)
	fmt.Println("Send:", string(data))

	conn.WriteMessage(websocket.TextMessage, data)

	return OK
}

func (t *rtcTest) processRecv(s *stage, conn *websocket.Conn,
	vars map[string]interface{}) int {

	msgchan := make(chan []byte)
	var timer *time.Timer
	if s.timeout > 0 {
		timer = time.NewTimer(s.timeout)
	} else {
		timer = time.NewTimer(24 * time.Hour)
	}

	go func() {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}

		fmt.Println("Recv:", string(msg))
		msgchan <- msg
	}()

	select {
	case msg := <-msgchan:
		timer.Stop()

		rm, ok := gjson.ParseBytes(msg).Value().(map[string]interface{})
		if !ok {
			log.Fatalln("Not json msg")
		}

		for k, v := range s.value {
			if rm[k] != v {
				log.Printf("para % check error, receive: %v, expect: %v",
					k, rm[k], v)
				return ERROR
			}
		}

		for k, v := range s.vars {
			if rm[v] != nil {
				vars[k] = rm[v]
			}
		}
	case <-timer.C:
		log.Println("wait for msg timeout")
		if s.option {
			return OK
		} else {
			return ERROR
		}
	}

	return OK
}

func (t *rtcTest) proceeIf(s *stage, vars map[string]interface{}) int {
	return OK
}

func (t *rtcTest) process(conn *websocket.Conn) bool {
	var ret int
	stid := 0
	vars := make(map[string]interface{})

	for {
		fmt.Println(vars)
		stage := t.stages[stid]
		switch stage.typ {
		case SEND:
			ret = t.processSend(stage, conn, vars)

		case RECV:
			ret = t.processRecv(stage, conn, vars)

		case IF:
			ret = t.proceeIf(stage, vars)

		case END:
			return true
		}

		switch ret {
		case ERROR:
			log.Print("Faile at stage", stid, stage)
			return false
		case OK:
			stid++
		default:
			stid = ret
		}

		if stid == len(t.stages) {
			return true
		}
	}
}

func wsCheckOrigin(r *http.Request) bool {
	return true
}

func (t *rtcTest) Handle(w http.ResponseWriter, req *http.Request) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  64 * 1024,
		WriteBufferSize: 64 * 1024,
		CheckOrigin:     wsCheckOrigin,
	}

	conn, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Println("Create Websocket server failed", err)
		return
	}

	if !t.process(conn) {
		t.failed++
	} else {
		t.successd++
	}

	fmt.Printf("failed: %d, successd: %d\n", t.failed, t.successd)
}

func (t *rtcTest) Run(url string, reqs uint64, conc uint64) {
	dialer := websocket.DefaultDialer

	conn, _, err := dialer.Dial(url, nil)

	if err != nil {
		log.Fatalln("Connect to Websocket server failed", err)
	}

	if !t.process(conn) {
		t.failed++
	} else {
		t.successd++
	}

	fmt.Printf("failed: %d, successd: %d\n", t.failed, t.successd)
}
