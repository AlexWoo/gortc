// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//
// go rtc script manager

package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"
	"time"

	simplejson "github.com/bitly/go-simplejson"
	"github.com/gorilla/websocket"
)

const (
	SEND = iota
	RECV
	TIMEOUT
)

type stage struct {
	typ     int
	timeout time.Duration
	value   interface{}
}

type rtcTest struct {
	stages   []*stage
	failed   uint32
	successd uint32
}

func NewRTCTest(file string) *rtcTest {
	rtctest := &rtcTest{}
	rtctest.compile(file)

	return rtctest
}

func getMapValue(v reflect.Value, key string) interface{} {
	if v.Kind() != reflect.Map {
		return nil
	}

	if v.MapIndex(reflect.ValueOf(key)).IsValid() {
		return v.MapIndex(reflect.ValueOf(key)).Interface()
	} else {
		return nil
	}
}

func (t *rtcTest) compile(file string) {
	f, err := os.Open(file)
	defer f.Close()
	if err != nil {
		log.Fatalln("open file", file, "error,", err)
	}

	json, err := simplejson.NewFromReader(f)
	if err != nil {
		log.Fatalln("parse script", file, "error,", err)
	}

	s := reflect.ValueOf(json.Interface())
	if s.Kind() != reflect.Slice {
		log.Fatalln("script format error, not slice")
	}

	for i := 0; i < s.Len(); i++ {
		stage := &stage{}

		st := s.Index(i).Elem()
		if st.Kind() != reflect.Map {
			log.Fatalln("script format error, stage not map")
		}

		switch getMapValue(st, "type") {
		case "send":
			stage.typ = SEND

			stage.value = getMapValue(st, "value")
			if stage.value == nil {
				log.Fatalln("Has no value in send")
			}

			timeout := getMapValue(st, "timeout")
			if timeout != nil {
				stage.timeout, err = time.ParseDuration(timeout.(string))
				if err != nil {
					log.Fatalln("timeout not time.Duration")
				}
			}

			t.stages = append(t.stages, stage)
		case "recv":
			stage.typ = RECV

			stage.value = getMapValue(st, "value")
			if stage.value == nil {
				log.Fatalln("Has no value in recv")
			}

			timeout := getMapValue(st, "timeout")
			if timeout != nil {
				stage.timeout, err = time.ParseDuration(timeout.(string))
				if err != nil {
					log.Fatalln("timeout not time.Duration")
				}
			}

			t.stages = append(t.stages, stage)
		default:
			log.Fatalln("unknown type", getMapValue(st, "type"))
		}
	}
}

func (t *rtcTest) process(conn *websocket.Conn) error {
	for i := 0; i < len(t.stages); i++ {
		stage := t.stages[i]
		switch stage.typ {
		case SEND:
			if stage.timeout > 0 {
				time.Sleep(stage.timeout)
			}
			conn.WriteJSON(stage.value)
		case RECV:
			msgchan := make(chan []byte)
			var timer *time.Timer
			if stage.timeout > 0 {
				timer = time.NewTimer(stage.timeout)
			} else {
				timer = time.NewTimer(24 * time.Hour)
			}

			go func(mc chan []byte) {
				_, msg, err := conn.ReadMessage()
				if err != nil {
					log.Println(err)
					msgchan <- msg
				}

				msgchan <- msg
			}(msgchan)

			select {
			case msg := <-msgchan:
				timer.Stop()

				fmt.Println(string(msg))

				if msg == nil {
					return errors.New("receive err msg")
				}

				json, err := simplejson.NewJson(msg)
				if err != nil {
					return errors.New("Not json msg")
				}

				rm, err := json.Map()
				if err != nil {
					return errors.New("json not map")
				}

				em := stage.value.(map[string]interface{})
				for k, v := range em {
					if rm[k] != v {
						return fmt.Errorf("para %s check error %v %v", k, rm[k], v)
					}
				}
			case <-timer.C:
				return fmt.Errorf("wait for msg timeout, stage %d", i)
			}
		}
	}

	return nil
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
		log.Fatalln("Create Websocket server failed", err)
	}

	err = t.process(conn)
	if err != nil {
		t.failed++
		log.Println(err)
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

	err = t.process(conn)
	if err != nil {
		t.failed++
		log.Println(err)
	} else {
		t.successd++
	}

	fmt.Printf("failed: %d, successd: %d\n", t.failed, t.successd)
}
