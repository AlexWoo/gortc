// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//
// go rtc test

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"rtclib"
	"runtime"
	"strconv"
)

const (
	UNKNOWN = iota
	UAC
	UAS
)

var (
	uatype      int
	scriptfile  string
	wsurl       string
	requests    uint64
	concurrency uint64
	addrport    string
	rtctest     *rtcTest
)

func usage() {
	fmt.Printf("usage: %s -h\n", os.Args[0])
	fmt.Printf("usage: %s -t uac -f scriptfile -u wsurl [-n requests] [-c concurrency]\n", os.Args[0])
	fmt.Printf("usage: %s -t uas -f scriptfile -l addrport\n", os.Args[0])
	os.Exit(1)
}

// UAC
func startUAC(wsurl string, requests uint64, concurrency uint64) {
	rtctest = NewRTCTest(scriptfile)
	if rtctest == nil {
		log.Fatalln("New RTC Test UAC failed")
	}

	rtctest.Run(wsurl, requests, concurrency)
}

// UAS
func startUAS(addrport string) {
	rtctest = NewRTCTest(scriptfile)
	if rtctest == nil {
		log.Fatalln("New RTC Test UAS failed")
	}

	http.HandleFunc("/", rtctest.Handle)
	log.Fatal(http.ListenAndServe(addrport, nil))
}

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	var err error

	opt := rtclib.NewOptParser()
	for opt.GetOpt("ht:f:u:n:c:l:]") {
		switch opt.Opt() {
		case 'h':
			usage()
		case 't': // uac or uas
			switch opt.OptVal() {
			case "uac":
				uatype = UAC
			case "uas":
				uatype = UAS
			default:
				log.Println("-t only accept value uac or uas")
				usage()
			}
		case 'f': // scriptfile
			scriptfile = opt.OptVal()
		case 'u': // wsurl
			wsurl = opt.OptVal()
		case 'n': // requests
			requests, err = strconv.ParseUint(opt.OptVal(), 10, 32)
			if err != nil {
				log.Println("-n only accept uint32 integer")
				usage()
			}
		case 'c': // concurrency
			concurrency, err = strconv.ParseUint(opt.OptVal(), 10, 32)
			if err != nil {
				log.Println("-c only accept uint32 integer")
				usage()
			}
		case 'l': // addrport to listen
			addrport = opt.OptVal()
		case '?':
			usage()
		}
	}

	if scriptfile == "" {
		log.Println("scriptfile must be set")
		usage()
	}

	switch uatype {
	case UAC:
		if wsurl == "" {
			log.Println("wsurl must be set for UAC")
			usage()
		}

		if requests == 0 {
			requests = 1
		}

		if concurrency == 0 {
			concurrency = 1
		}

		startUAC(wsurl, requests, concurrency)
	case UAS:
		if addrport == "" {
			log.Println("addrport must be set for UAS")
			usage()
		}

		startUAS(addrport)
	default:
		log.Println("uatype must be set")
		usage()
	}
}
