// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//
// Base Type

package rtclib

const (
	UAS = iota
	UAC
)

// Base Conn type
type Conn interface {
	// Connect to server as client
	Dial()

	// Accept connection as server
	Accept()

	// Send data
	Send(data []byte)

	// Close connection
	Close()
}
