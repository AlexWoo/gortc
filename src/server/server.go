package main

import (
	"fmt"
	"net/http"
	"rtclib"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

var conns = make(map[string]*rtclib.WSConn)

func wsCheckOrigin(r *http.Request) bool {
	return true
}

func send(c rtclib.Conn) {
	for i := 0; i < 100000; i++ {
		c.Send([]byte(strconv.Itoa(i)))
		time.Sleep(100 * time.Millisecond)
	}
}

func handler(c rtclib.Conn, data []byte) {
	fmt.Println(string(data))
}

func recv(w http.ResponseWriter, req *http.Request) {
	upgrade := websocket.Upgrader{
		ReadBufferSize:  64 * 1024,
		WriteBufferSize: 64 * 1024,
		CheckOrigin:     wsCheckOrigin,
	}

	c, err := upgrade.Upgrade(w, req, nil)
	if err != nil {
		fmt.Println("websocket server: ", err)
		return
	}

	conn := rtclib.NewWSConn("a", "", rtclib.UAS, 3*time.Second, 1024, handler)
	if conns["a"] == nil {
		go send(conn)
		conns["a"] = conn
	}

	conn.Accept(c)
}

func main() {
	http.HandleFunc("/", recv)
	http.ListenAndServe(":8080", nil)
}
