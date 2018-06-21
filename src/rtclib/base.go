// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//
// Base Type

package rtclib

const (
	UAS = iota
	UAC
)

type Size_t uint64

// Base Conn type
type Conn interface {
	// Accept connection as server
	Accept(data interface{})

	// Send data
	Send(data []byte)

	// Close connection
	Close()
}
