// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//
// RTC Config

package rtclib

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-ini/ini"
)

type Size_t uint64
type Msec_t uint64

func confValue(key string, s *ini.Section) string {
	return s.Key(key).Value()
}

func strToBoolean(value string) (bool, bool) {
	if strings.ToLower(value) == "true" {
		return true, true
	} else if strings.ToLower(value) == "false" {
		return false, true
	}

	return false, false
}

func defaultBoolean(value string) bool {
	ret, ok := strToBoolean(value)

	if ok {
		return ret
	}

	return false
}

func confBoolean(value string, defaultVal bool) bool {
	ret, ok := strToBoolean(value)

	if ok {
		return ret
	}

	return defaultVal
}

func confString(value string, defaultVal string) string {
	if value == "" {
		return defaultVal
	}

	return value
}

func strToUint64(value string) (uint64, bool) {
	ret, err := strconv.ParseUint(value, 10, 64)

	return ret, err == nil
}

func defaultUint64(value string) uint64 {
	ret, ok := strToUint64(value)

	if ok {
		return ret
	}

	return 0
}

func confUint64(value string, defaultVal uint64) uint64 {
	ret, ok := strToUint64(value)

	if ok {
		return ret
	}

	return defaultVal
}

func strToInt64(value string) (int64, bool) {
	ret, err := strconv.ParseInt(value, 10, 64)

	return ret, err == nil
}

func defaultInt64(value string) int64 {
	ret, ok := strToInt64(value)

	if ok {
		return ret
	}

	return 0
}

func confInt64(value string, defaultVal int64) int64 {
	ret, ok := strToInt64(value)

	if ok {
		return ret
	}

	return defaultVal
}

func strToSize(value string) (Size_t, bool) {
	if value == "" {
		return 0, false
	}

	var data []byte = []byte(value)
	length := len(data)

	end := data[length-1]
	var base uint64 = 1
	switch end {
	case 'K':
		fallthrough
	case 'k':
		length--
		base = base * 1024
	case 'M':
		fallthrough
	case 'm':
		length--
		base = base * 1024 * 1024
	case 'G':
		fallthrough
	case 'g':
		length--
		base = base * 1024 * 1024 * 1024
	}

	newStr := string(data[:length])

	ret, ok := strToUint64(newStr)
	if ok {
		ret = ret * base
		return Size_t(ret), true
	}

	return 0, false
}

func defaultSize(value string) Size_t {
	ret, ok := strToSize(value)

	if ok {
		return ret
	}

	return 0
}

func confSize(value string, defaultVal Size_t) Size_t {
	ret, ok := strToSize(value)

	if ok {
		return ret
	}

	return defaultVal
}

func strToMsec(value string) (Msec_t, bool) {
	if value == "" {
		return 0, false
	}

	var data []byte = []byte(value)
	length := len(data)

	end := data[length-1]
	var base uint64 = 1
	switch end {
	case 'h':
		length--
		base = base * 60 * 60 * 1000
	case 'm':
		length--
		base = base * 60 * 1000
	case 's':
		length--
		if data[length-1] == 'm' { //msec
			length--
		} else { //sec
			base = base * 1000
		}
	}

	newStr := string(data[:length])

	ret, ok := strToUint64(newStr)
	if ok {
		ret = ret * base
		return Msec_t(ret), true
	}

	return 0, false
}

func defaultMsec(value string) Msec_t {
	ret, ok := strToMsec(value)

	if ok {
		return ret
	}

	return 0
}

func confMsec(value string, defaultVal Msec_t) Msec_t {
	ret, ok := strToMsec(value)

	if ok {
		return ret
	}

	return defaultVal
}

func ConfEnum(e map[string]int, value string, defaultVal int) int {
	ret, ok := e[value]

	if ok {
		return ret
	}

	return defaultVal
}

func Config(f *ini.File, secName string, it interface{}) bool {
	if secName == "DEFAULT" {
		secName = ""
	}

	s := f.Section(secName)

	t := reflect.TypeOf(it).Elem()
	v := reflect.ValueOf(it).Elem()
	n := t.NumField()

	for i := 0; i < n; i++ {
		field := t.Field(i)
		value := v.Field(i)
		fn := field.Name
		ft := field.Type.Name()
		fd := field.Tag.Get("default")

		switch ft {
		case "bool":
			confV := confValue(strings.ToLower(fn), s)
			value.SetBool(confBoolean(confV, defaultBoolean(fd)))
		case "string":
			confV := confValue(strings.ToLower(fn), s)
			value.SetString(confString(confV, fd))
		case "uint64":
			confV := confValue(strings.ToLower(fn), s)
			value.SetUint(confUint64(confV, defaultUint64(fd)))
		case "int64":
			confV := confValue(strings.ToLower(fn), s)
			value.SetInt(confInt64(confV, defaultInt64(fd)))
		case "Size_t":
			confV := confValue(strings.ToLower(fn), s)
			value.SetUint(uint64(confSize(confV, defaultSize(fd))))
		case "Msec_t":
			confV := confValue(strings.ToLower(fn), s)
			value.SetUint(uint64(confMsec(confV, defaultMsec(fd))))
		default:
			fmt.Printf("Unsuppoted config, secName: %s, name: %s, type: %s\n",
				secName, fn, ft)
			return false
		}
	}

	return true
}
