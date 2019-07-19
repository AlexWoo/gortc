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
	"strconv"
	"syscall"

	"github.com/alexwoo/golib"
)

var (
	pidfile = rtclib.FullPath(".gortc.pid")
)

func usage() {
	fmt.Printf("usage: %s -h\n", os.Args[0])
	fmt.Printf("usage: %s [-d]\n", os.Args[0])
	fmt.Printf("    -d    start backgroud\n")
	fmt.Printf("    -s quit|stop|reopen|reload\n")
	fmt.Printf("          quit: gortc quit directly\n")
	fmt.Printf("          stop: gortc quit gracefully\n")
	fmt.Printf("          reopen: gortc reopen logs\n")
	fmt.Printf("          reload: gortc reload config\n")
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
	_, err := os.StartProcess(rtclib.FullPath("bin/gortc"), args,
		&os.ProcAttr{Files: []*os.File{os.Stdin, os.Stdout, os.Stderr}})

	if err != nil {
		log.Println("Daemon start failed:", err)
	}

	os.Exit(0)
}

func signal(cmd string) {
	pid := readPIDFile()
	if pid == -1 {
		fmt.Println("read pidfile", pidfile, "failed")
		os.Exit(-1)
	}

	switch cmd {
	case "stop":
		syscall.Kill(pid, syscall.SIGTERM)
	case "quit":
		syscall.Kill(pid, syscall.SIGQUIT)
	case "reopen":
		syscall.Kill(pid, syscall.SIGUSR1)
	case "reload":
		syscall.Kill(pid, syscall.SIGHUP)
	default:
		fmt.Println("Unknown command for gortc -s", cmd)
		os.Exit(-1)
	}

	os.Exit(0)
}

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func writePIDFile() {
	f, err := os.OpenFile(pidfile, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0600)
	if err != nil {
		fmt.Println("Write pid file failed", err)
		os.Exit(-1)
	}

	defer f.Close()

	f.WriteString(fmt.Sprintf("%d", os.Getpid()))
}

func readPIDFile() int {
	f, err := os.Open(pidfile)
	if err != nil {
		return -1
	}

	defer f.Close()

	b := make([]byte, 16)
	l, err := f.Read(b)
	if err != nil {
		return -1
	}

	if pid, err := strconv.Atoi(string(b[:l])); err != nil {
		fmt.Println(err)
		return -1
	} else {
		return pid
	}
}

func unlinkPIDFile() {
	os.Remove(pidfile)
}

func main() {
	opt := golib.NewOptParser()
	for opt.GetOpt("hds:") {
		switch opt.Opt() {
		case 'h':
			usage()
		case 'd':
			daemon()
		case 's':
			signal(opt.OptVal())
		case '?':
			usage()
		}
	}

	writePIDFile()

	ms := golib.NewModules()

	ms.AddModule("main", &mainModule{})
	ms.AddModule("apiserver", apiServerInstance())
	ms.AddModule("apimanager", apimInstance())
	ms.AddModule("distribute", distInstance())
	ms.AddModule("rtcserver", rtcServerInstance())
	ms.AddModule("slpmanager", slpmInstance())

	ms.Start()

	unlinkPIDFile()
}
