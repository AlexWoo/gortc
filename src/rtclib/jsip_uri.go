// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//

// JSIP URI

package rtclib

import (
	"errors"
	"strconv"
	"strings"
)

// JSIPUri Para for Unmarshal para string as para1=value or para2
type JSIPUriPara struct {
	Key   string
	Value string
}

// Unmarshal JSIPUri Para string to JSIPURIPara
func NewJSIPUriPara(raw string) (*JSIPUriPara, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, errors.New("Null para")
	}

	split := strings.SplitN(raw, "=", 2)

	p := &JSIPUriPara{
		Key: split[0],
	}

	if len(split) == 2 {
		p.Value = split[1]
	}

	return p, nil
}

// return JSIPUriPara as string
func (p *JSIPUriPara) String() string {
	output := p.Key
	if p.Value != "" {
		output += "=" + p.Value
	}

	return output
}

// JSIPUri Hostport for Unmarshal hostport string as host[:port]
type JSIPUriHostport struct {
	Host string
	Port uint16
}

// Unmarshal hostport string to JSIPUriHostport
func NewJSIPUriHostport(raw string) (*JSIPUriHostport, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, errors.New("Null hostport")
	}

	// now not support ipv6
	split := strings.SplitN(raw, ":", 2)

	hp := &JSIPUriHostport{
		Host: split[0],
	}

	if len(split) == 1 {
		return hp, nil
	}

	port, err := strconv.Atoi(split[1])
	if err != nil {
		return nil, errors.New("Port not int")
	}

	if port < 0 || port > 65535 {
		return nil, errors.New("Invalid port")
	}

	hp.Port = uint16(port)

	return hp, nil
}

// return JSIPUriHostport as string
func (hp *JSIPUriHostport) String() string {
	output := hp.Host
	if hp.Port != 0 {
		output += ":" + strconv.Itoa(int(hp.Port))
	}

	return output
}

// JSIPUri for Unmarshal uri string as [user@]host[:port][;para1=value][;para2]
type JSIPUri struct {
	User     string
	Hostport *JSIPUriHostport
	Paras    map[string]string

	paras []*JSIPUriPara
}

// Unmarshal uri string to JSIPUri
func NewJSIPUri(raw string) (*JSIPUri, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, errors.New("Null uri")
	}

	u := &JSIPUri{
		Paras: map[string]string{},
		paras: []*JSIPUriPara{},
	}

	split := strings.SplitN(raw, ";", 2)
	if len(split) == 2 {
		parastr := split[1]

		paras := strings.Split(parastr, ";")
		for _, para := range paras {
			jp, err := NewJSIPUriPara(para)
			if err == nil {
				u.paras = append(u.paras, jp)
				u.Paras[jp.Key] = jp.Value
			}
		}
	}

	hostpart := split[0]
	split = strings.Split(hostpart, "@")
	if len(split) > 2 {
		return nil, errors.New("Error hostpart")
	}

	hostport := ""
	if len(split) == 1 {
		hostport = split[0]
	} else {
		u.User = split[0]
		hostport = split[1]
	}

	hp, err := NewJSIPUriHostport(hostport)
	if err != nil {
		return nil, err
	}

	u.Hostport = hp

	return u, nil
}

// return JSIPUri Hostpart as string: [user@]host[:port]
func (u *JSIPUri) HostpartString() string {
	output := u.Hostport.String()
	if u.User != "" {
		output = u.User + "@" + output
	}

	return output
}

// return JSIPUri Hostport as string: host[:port]
func (u *JSIPUri) HostportString() string {
	return u.Hostport.String()
}

// return JSIPUri UserHost as string: [user@]host
func (u *JSIPUri) UserHostString() string {
	output := u.Hostport.Host
	if u.User != "" {
		output = u.User + "@" + output
	}

	return output
}

// return JSIPUri para string: [;para1=value][;para2]
func (u *JSIPUri) ParasString() string {
	output := ""
	for _, p := range u.paras {
		output += ";" + p.String()
	}

	return output
}
