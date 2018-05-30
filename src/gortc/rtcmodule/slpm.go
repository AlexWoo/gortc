// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//
// RTC Service Logical Processor manager

package rtcmodule

import (
	"fmt"
	"gortc/apimodule"
	"os"
	"plugin"
	"rtclib"
	"time"

	simplejson "github.com/bitly/go-simplejson"
)

type slpPlugin struct {
	name     string
	used     uint64
	using    uint64
	file     string
	time     time.Time
	ctx      *rtclib.SlpCtx
	instance func(task *rtclib.Task) rtclib.SLP
}

type SLPM struct {
	slps    map[string]*slpPlugin
	slpconf string
	slpdir  string
	plugins map[string]string
}

var slpm = &SLPM{
	slps:    make(map[string]*slpPlugin),
	plugins: make(map[string]string),
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

	defer func() {
		if err := recover(); err != nil {
			LogError("load %s %s failed: %v", name, path, err)
			if slpm.plugins[name] == "" {
				delete(slpm.plugins, name)
				updateSLPFile()
			}
		}
	}()

	slp.ctx = new(rtclib.SlpCtx)
	slp.instance = v.(func(task *rtclib.Task) rtclib.SLP)
	slp.time = time.Now()
	slpm.slps[name] = slp

	return true
}

func initSLPM() bool {
	slpm.slpconf = rtclib.RTCPATH + "/conf/.slps"
	slpm.slpdir = rtclib.RTCPATH + "/slp/"

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
		LogError("slp file %s format error: %v", slpm.slpconf, err)
		return false
	}

	for name, _ := range j {
		path, err := json.Get(name).String()
		if err != nil {
			LogError("slp %s format error: %v", name, err)
			return false
		}

		if !slpLoad(name, path) {
			return false
		}
		slpm.plugins[name] = path
	}

	apimodule.AddInternalAPI("slpm.v1", Slpmv1)

	return true
}

func updateSLPFile() error {
	json := simplejson.New()
	for k, v := range slpm.plugins {
		json.Set(k, v)
	}

	f, err := os.OpenFile(slpm.slpconf, os.O_TRUNC|os.O_WRONLY, 0644)
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

	slpm.plugins[name] = slpFile

	err := updateSLPFile()
	if err != nil {
		return fmt.Sprintf("Update slp conf file %s error: %v\n",
			slpm.slpconf, err)
	}

	return fmt.Sprintf("Load SLP %s %s successd\n", name, slpFile)
}

func delSLP(name string) string {
	delete(slpm.plugins, name)
	delete(slpm.slps, name)

	err := updateSLPFile()
	if err != nil {
		return fmt.Sprintf("Update slp conf file %s error: %v\n",
			slpm.slpconf, err)
	}

	return fmt.Sprintf("Delete SLP %s successd\n", name)
}

func getSLP(t *rtclib.Task) rtclib.SLP {
	p := slpm.slps[t.Name]
	if p == nil {
		LogError("SLP %s not exist", t.Name)
		return nil
	}
	p.using++

	t.Ctx = p.ctx
	return p.instance(t)
}

func endSLP(t *rtclib.Task) {
	p := slpm.slps[t.Name]
	if p == nil { // SLP has been deleted
		return
	}

	p.using--
	p.used++
}

func listSLP() string {
	ret := "slp\t\tused\t\tusing\t\tfile\t\ttime\n"
	ret += "------------------------------------------------------------\n"
	for _, v := range slpm.slps {
		ret += fmt.Sprintf("%s\t%d\t%d\t%s\t%s\n", v.name, v.used, v.using,
			v.file, v.time.Format("2006-01-02 15:04:05.000"))
	}
	ret += "------------------------------------------------------------\n"

	return ret
}
