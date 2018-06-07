package main

import (
	"fmt"
	"rtclib"
	"strconv"
	"time"
)

func send(c rtclib.Conn) {
	for i := 0; i < 100000; i++ {
		c.Send([]byte(strconv.Itoa(i)))
		time.Sleep(10 * time.Millisecond)
	}
}

func handler(c rtclib.Conn, data []byte) {
	fmt.Println(string(data))
}

func main() {
	conn := rtclib.NewWSConn("a", "ws://127.0.0.1:8080", rtclib.UAC,
		3*time.Second, 1024, handler)
	if conn == nil {
		return
	}

	wait := make(chan bool)

	conn.Dial()

	go send(conn)

	<-wait
}
