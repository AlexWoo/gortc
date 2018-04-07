// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//
// option parser

package rtclib

import (
	"os"
)

type OptParser struct {
	opt    byte
	optval string
	optidx int
}

func NewOptParser() *OptParser {
	return &OptParser{optidx: 1}
}

func (opt *OptParser) OptVal() string {
	return opt.optval
}

func (opt *OptParser) Opt() byte {
	return opt.opt
}

func (opt *OptParser) GetOpt(optstring string) bool {
	if opt.optidx >= len(os.Args) {
		return false
	}

	optstr := []byte(optstring)
	arg := []byte(os.Args[opt.optidx])
	if arg[0] != '-' { // not an option
		goto failed
	}

	if !((arg[1] == '?') || (arg[1] >= '0' && arg[1] <= '9') ||
		(arg[1] >= 'a' && arg[1] <= 'z') || (arg[1] >= 'A' && arg[1] <= 'Z')) {

		goto failed
	}

	for i := 0; i < len(optstr); i++ {
		if optstr[i] == arg[1] {
			if i+1 == len(optstr) || optstr[i+1] != ':' { // no argument
				if len(arg) != 2 { // has argument
					goto failed
				}

				opt.opt = arg[1]
				opt.optval = ""
				opt.optidx++
				return true
			} else { // has argument
				if len(arg) != 2 { // optval stick with option
					opt.opt = arg[1]
					opt.optval = string(arg[2:])
					opt.optidx++
					return true
				}

				if opt.optidx+1 == len(os.Args) {
					goto failed
				}

				opt.opt = arg[1]
				opt.optval = os.Args[opt.optidx+1]
				opt.optidx += 2
				return true
			}
		}
	}

failed:
	opt.opt = '?'
	opt.optval = ""
	opt.optidx++

	return true
}
