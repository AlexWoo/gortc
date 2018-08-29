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

type APIM struct {
	apis    map[string]*apiPlugin
	apiconf string
	apidir  string
	plugins map[string]string
}

var apim = &APIM{
	apis:    make(map[string]*apiPlugin),
	plugins: make(map[string]string),
}

func addInternalAPI(name string, instance func() rtclib.API) {
	p := &apiPlugin{
		name:     name,
		instance: instance,
	}

	apim.apis[name] = p
}

func apiLoad(name string, apiFile string) bool {
	api := &apiPlugin{
		name: name,
		file: apiFile,
	}
	path := apim.apidir + apiFile

	p, err := plugin.Open(path)
	if err != nil {
		apis.LogError("open api plugin error: %v", err)
		return false
	}

	v, err := p.Lookup("APIInstance")
	if err != nil {
		apis.LogError("find api plugin entry error: %v", err)
		return false
	}

	defer func() {
		if err := recover(); err != nil {
			apis.LogError("load %s %s failed: %v", name, path, err)
			if apim.plugins[name] == "" {
				delete(apim.plugins, name)
				updateAPIFile()
			}
		}
	}()

	api.instance = v.(func() rtclib.API)
	apim.apis[name] = api

	return true
}

func initAPIM() bool {
	apim.apiconf = rtclib.FullPath("conf/.apis")
	apim.apidir = rtclib.FullPath("api/")

	f, err := os.Open(apim.apiconf)
	defer f.Close()
	if err != nil {
		apis.LogError("open file %s failed: %v", apim.apiconf, err)
		return false
	}

	json, _ := ioutil.ReadAll(f)
	if !gjson.ValidBytes(json) {
		apis.LogError("parse file %s failed: %v", apim.apiconf, err)
		return false
	}

	j, ok := gjson.ParseBytes(json).Value().(map[string]interface{})
	if !ok {
		apis.LogError("api file %s format error: %v", apim.apiconf, err)
		return false
	}

	for name, v := range j {
		path, ok := v.(string)
		if !ok {
			apis.LogError("api %s format error: %v", name, err)
			return false
		}

		if !apiLoad(name, path) {
			return false
		}
		apim.plugins[name] = path
	}

	addInternalAPI("apim.v1", Apimv1)

	return true
}

func updateAPIFile() error {
	f, err := os.OpenFile(apim.apiconf, os.O_TRUNC|os.O_WRONLY, 0644)
	defer f.Close()
	if err != nil {
		return fmt.Errorf("open file %s failed: %v", apim.apiconf, err)
	}

	d, _ := json.Marshal(apim.plugins)

	_, err = f.Write(d)
	if err != nil {
		return fmt.Errorf("write file %s failed: %v", apim.apiconf, err)
	}

	return nil
}

func addAPI(name string, apiFile string) string {
	if !apiLoad(name, apiFile) {
		return fmt.Sprintf("Load API %s %s failed\n", name, apiFile)
	}

	apim.plugins[name] = apiFile

	err := updateAPIFile()
	if err != nil {
		return fmt.Sprintf("Update api conf file %s error: %v\n",
			apim.apiconf, err)
	}

	return fmt.Sprintf("Load API %s %s successd\n", name, apiFile)
}

func delAPI(name string) string {
	if apim.plugins[name] == "" {
		return fmt.Sprintf("Cannot delete api %s\n", name)
	}

	delete(apim.plugins, name)
	delete(apim.apis, name)

	err := updateAPIFile()
	if err != nil {
		return fmt.Sprintf("Update api conf file %s error: %v\n",
			apim.apiconf, err)
	}

	return fmt.Sprintf("Delete API %s successd\n", name)
}

func getAPI(name string) rtclib.API {
	p := apim.apis[name]
	if p == nil {
		return nil
	}

	return p.instance()
}

func listAPI() string {
	ret := "api\t\tfile\n"
	ret += "------------------------------------------------------------\n"
	for _, v := range apim.apis {
		ret += fmt.Sprintf("%s\t%s\n", v.name, v.file)
	}
	ret += "------------------------------------------------------------\n"

	return ret
}
