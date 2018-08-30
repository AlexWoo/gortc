// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//
// API manager

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"plugin"
	"rtclib"

	"github.com/tidwall/gjson"
)

type apiPlugin struct {
	name     string
	file     string
	instance func() rtclib.API
}

type apim struct {
	apis    map[string]*apiPlugin
	apiconf string
	apidir  string
	plugins map[string]string
}

var am *apim

func apimInstance() *apim {
	if am != nil {
		return am
	}

	am = &apim{
		apis:    make(map[string]*apiPlugin),
		plugins: make(map[string]string),
	}

	return am
}

// for module interface

func (m *apim) PreInit() error {
	return nil
}

func (m *apim) Init() error {
	m.apiconf = rtclib.FullPath("conf/.apis")
	m.apidir = rtclib.FullPath("api/")

	f, err := os.Open(m.apiconf)
	defer f.Close()
	if err != nil {
		return fmt.Errorf("open file %s failed: %v", m.apiconf, err)
	}

	json, _ := ioutil.ReadAll(f)
	if !gjson.ValidBytes(json) {
		return fmt.Errorf("parse file %s failed: %v", m.apiconf, err)
	}

	j, ok := gjson.ParseBytes(json).Value().(map[string]interface{})
	if !ok {
		return fmt.Errorf("api file %s format error: %v", m.apiconf, err)
	}

	for name, v := range j {
		path, ok := v.(string)
		if !ok {
			return fmt.Errorf("api %s format error: %v", name, err)
		}

		if err := m.apiLoad(name, path); err != nil {
			return err
		}
		m.plugins[name] = path
	}

	return nil
}

func (m *apim) PreMainloop() error {
	m.addInternalAPI("apim.v1", Apimv1)

	return nil
}

func (m *apim) Mainloop() {
}

func (m *apim) Reload() error {
	return nil
}

func (m *apim) Reopen() error {
	return nil
}

func (m *apim) Exit() {
}

// internal interalface

func (m *apim) addInternalAPI(name string, instance func() rtclib.API) {
	p := &apiPlugin{
		name:     name,
		instance: instance,
	}

	m.apis[name] = p
}

func (m *apim) updateAPIFile() error {
	f, err := os.OpenFile(m.apiconf, os.O_TRUNC|os.O_WRONLY, 0644)
	defer f.Close()
	if err != nil {
		return fmt.Errorf("open file %s failed: %v", m.apiconf, err)
	}

	d, _ := json.Marshal(m.plugins)

	_, err = f.Write(d)
	if err != nil {
		return fmt.Errorf("write file %s failed: %v", m.apiconf, err)
	}

	return nil
}

func (m *apim) apiLoad(name string, apiFile string) error {
	api := &apiPlugin{
		name: name,
		file: apiFile,
	}
	path := m.apidir + apiFile

	p, err := plugin.Open(path)
	if err != nil {
		return fmt.Errorf("open api plugin error: %v", err)
	}

	v, err := p.Lookup("APIInstance")
	if err != nil {
		return fmt.Errorf("find api plugin entry error: %v", err)
	}

	defer func() {
		if err := recover(); err != nil {
			apis.LogError("load %s %s failed: %v", name, path, err)
			if m.plugins[name] == "" {
				delete(m.plugins, name)
				m.updateAPIFile()
			}
		}
	}()

	api.instance = v.(func() rtclib.API)
	m.apis[name] = api

	return nil
}

func (m *apim) addAPI(name string, apiFile string) string {
	if err := m.apiLoad(name, apiFile); err != nil {
		return fmt.Sprintf("Load API %s %s failed, %v\n", name, apiFile, err)
	}

	m.plugins[name] = apiFile

	err := m.updateAPIFile()
	if err != nil {
		return fmt.Sprintf("Update api conf file %s error: %v\n",
			m.apiconf, err)
	}

	return fmt.Sprintf("Load API %s %s successd\n", name, apiFile)
}

func (m *apim) delAPI(name string) string {
	if m.plugins[name] == "" {
		return fmt.Sprintf("Cannot delete api %s\n", name)
	}

	delete(m.plugins, name)
	delete(m.apis, name)

	err := m.updateAPIFile()
	if err != nil {
		return fmt.Sprintf("Update api conf file %s error: %v\n",
			m.apiconf, err)
	}

	return fmt.Sprintf("Delete API %s successd\n", name)
}

func (m *apim) getAPI(name string) rtclib.API {
	p := m.apis[name]
	if p == nil {
		return nil
	}

	return p.instance()
}

func (m *apim) listAPI() string {
	ret := "api\t\tfile\n"
	ret += "------------------------------------------------------------\n"
	for _, v := range m.apis {
		ret += fmt.Sprintf("%s\t%s\n", v.name, v.file)
	}
	ret += "------------------------------------------------------------\n"

	return ret
}
