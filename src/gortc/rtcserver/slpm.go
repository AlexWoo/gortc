// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//
// RTC Service Logical Processor manager

package rtcserver

import (
	"fmt"
	"os"
	"plugin"
	"time"

	simplejson "github.com/bitly/go-simplejson"
)

type SLP interface {
	Process(jsip *JSIP) int
}

type slpPlugin struct {
	name     string
	used     uint64
	using    uint64
	file     string
	time     time.Time
	instance func(task *Task) SLP
}

type SLPM struct {
	selector map[string]*slpPlugin
	slpconf  string
	slpdir   string
	slp      map[string]string
}

var slpm = &SLPM{
	selector: make(map[string]*slpPlugin),
	slp:      make(map[string]string),
}

func slpLoad(name string, slpFile string) bool {
	slp := &slpPlugin{name: name, file: slpFile}
	path := slpm.slpdir + slpFile

	p, err := plugin.Open(path)
	if err != nil {
		LogError("open slp plugin error: %v", err)
		return false
	}

	v, err := p.Lookup("GetInstance")
	if err != nil {
		LogError("find slp plugin entry error: %v", err)
		return false
	}

	slp.instance = v.(func(task *Task) SLP)

	slpm.selector[name] = slp

	return true
}

func initSelector() bool {
	slpm.slpconf = rtcServerModule.rtcPath + rtcServerModule.config.SLPSelector
	slpm.slpdir = rtcServerModule.rtcPath + "/slp/"

	f, err := os.Open(slpm.slpconf)
	defer f.Close()
	if err != nil {
		LogError("open file %s failed: %v", slpm.slpconf, err)
		return false
	}

	json, err := simplejson.NewFromReader(f)
	if err != nil {
		LogError("parse file %s failed: %v", slpm.slpconf, err)
		return false
	}

	j, err := json.Map()
	if err != nil {
		LogError("selector file %s format error: %v", slpm.slpconf, err)
		return false
	}

	for name, _ := range j {
		path, err := json.Get(name).String()
		if err != nil {
			LogError("select %s format error: %v", name, err)
			return false
		}

		if !slpLoad(name, path) {
			return false
		}
		slpm.slp[name] = path
	}

	return true
}

func updateSLPFile() error {
	json := simplejson.New()
	for k, v := range slpm.slp {
		json.Set(k, v)
	}

	f, err := os.Open(slpm.slpconf)
	defer f.Close()
	if err != nil {
		return fmt.Errorf("open file %s failed: %v", slpm.slpconf, err)
	}

	d, _ := json.MarshalJSON()

	_, err = f.Write(d)
	if err != nil {
		return fmt.Errorf("write file %s failed: %v", slpm.slpconf, err)
	}

	return nil
}

func addSLP(name string, slpFile string) string {
	if !slpLoad(name, slpFile) {
		return fmt.Sprintf("Load SLP %s %s failed\n", name, slpFile)
	}

	slpm.slp[name] = slpFile

	err := updateSLPFile()
	if err != nil {
		return fmt.Sprintf("Update slp conf file %s error: %v\n",
			slpm.slpconf, err)
	}

	return fmt.Sprintf("Load SLP %s %s successd\n", name, slpFile)
}

func delSLP(name string) string {
	delete(slpm.slp, name)
	delete(slpm.selector, name)

	err := updateSLPFile()
	if err != nil {
		return fmt.Sprintf("Update slp conf file %s error: %v\n",
			slpm.slpconf, err)
	}

	return fmt.Sprintf("Delete SLP %s successd\n", name)
}

func getSLP(t *Task) SLP {
	p := slpm.selector[t.name]
	if p == nil {
		LogError("SLP %s not exist", t.name)
		return nil
	}
	p.using++

	return p.instance(t)
}

func endSLP(t *Task) {
	p := slpm.selector[t.name]
	if p == nil { // SLP has been deleted
		return
	}

	p.using--
	p.used++
}

func listSLP() string {
	ret := "slp\t\tused\t\tusing\t\tfile\t\ttime\n"
	ret += "------------------------------------------------------------\n"
	for _, v := range slpm.selector {
		ret += fmt.Sprintf("%s\t%d\t%d\t%s\t%s\n", v.name, v.used, v.using,
			v.file, v.time.Format("2006-01-02 15:04:05.000"))
	}
	ret += "------------------------------------------------------------\n"

	return ret
}