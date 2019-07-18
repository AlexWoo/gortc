// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//
// main func entry

package main

import (
	"fmt"
	"log"
	"os"
	"rtclib"
	"runtime"
	"syscall"

	"github.com/alexwoo/golib"
)

var (
	rtcpath = "/usr/local/gortc/"
)

func usage() {
	fmt.Printf("usage: %s -h\n", os.Args[0])
	fmt.Printf("usage: %s [-d]\n", os.Args[0])
	fmt.Printf("    -d    start backgroud\n")
	os.Exit(1)
}

func daemon() {
	if syscall.Getppid() == 1 { // already daemon
		f, err := os.OpenFile("/dev/null", os.O_RDWR, 0)
		if err != nil {
			fmt.Println("open /dev/null failed")
			os.Exit(-1)
		}

		fd := f.Fd()
		syscall.Dup2(int(fd), int(os.Stdin.Fd()))
		syscall.Dup2(int(fd), int(os.Stdout.Fd()))
		syscall.Dup2(int(fd), int(os.Stderr.Fd()))

		return
	}

	args := append([]string{os.Args[0]}, os.Args[1:]...)
	_, err := os.StartProcess(os.Args[0], args,
		&os.ProcAttr{Files: []*os.File{os.Stdin, os.Stdout, os.Stderr}})

	if err != nil {
		log.Println("Daemon start failed:", err)
	}

	os.Exit(0)
}

func init() {
	rtclib.RTCPATH = rtcpath
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	opt := golib.NewOptParser()
	for opt.GetOpt("hd") {
		switch opt.Opt() {
		case 'h':
			usage()
		case 'd':
			daemon()
		case '?':
			usage()
		}
	}

	ms := golib.NewModules()

	ms.AddModule("main", &mainModule{})
	ms.AddModule("apiserver", apiServerInstance())
	ms.AddModule("apimanager", apimInstance())
	ms.AddModule("distribute", distInstance())
	ms.AddModule("rtcserver", rtcServerInstance())
	ms.AddModule("slpmanager", slpmInstance())

	ms.Start()
}
