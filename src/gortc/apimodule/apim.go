// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//
// API manager

package apimodule

import (
	"fmt"
	"net/http"
	"os"
	"plugin"
	"rtclib"

	simplejson "github.com/bitly/go-simplejson"
)

type API interface {
	Get(req *http.Request, paras string) (int,
		*map[string]string, interface{}, *map[int]RespCode)
	Post(req *http.Request, paras string) (int,
		*map[string]string, interface{}, *map[int]RespCode)
	Delete(req *http.Request, paras string) (int,
		*map[string]string, interface{}, *map[int]RespCode)
}

type apiPlugin struct {
	name     string
	file     string
	instance func() API
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

func AddInternalAPI(name string, instance func() API) {
	p := &apiPlugin{name: name, instance: instance}

	apim.apis[name] = p
}

func apiLoad(name string, apiFile string) bool {
	api := &apiPlugin{name: name, file: apiFile}
	path := apim.apidir + apiFile

	p, err := plugin.Open(path)
	if err != nil {
		LogError("open api plugin error: %v", err)
		return false
	}

	v, err := p.Lookup("GetInstance")
	if err != nil {
		LogError("find api plugin entry error: %v", err)
		return false
	}

	defer func() {
		if err := recover(); err != nil {
			LogError("load %s %s failed: %v", name, path, err)
			if apim.plugins[name] == "" {
				delete(apim.plugins, name)
				updateAPIFile()
			}
		}
	}()

	api.instance = v.(func() API)
	apim.apis[name] = api

	return true
}

func initAPIM() bool {
	apim.apiconf = rtclib.RTCPATH + "/conf/.apis"
	apim.apidir = rtclib.RTCPATH + "/api/"

	f, err := os.Open(apim.apiconf)
	defer f.Close()
	if err != nil {
		LogError("open file %s failed: %v", apim.apiconf, err)
		return false
	}

	json, err := simplejson.NewFromReader(f)
	if err != nil {
		LogError("parse file %s failed: %v", apim.apiconf, err)
		return false
	}

	j, err := json.Map()
	if err != nil {
		LogError("api file %s format error: %v", apim.apiconf, err)
		return false
	}

	for name, _ := range j {
		path, err := json.Get(name).String()
		if err != nil {
			LogError("api %s format error: %v", name, err)
			return false
		}

		if !apiLoad(name, path) {
			return false
		}
		apim.plugins[name] = path
	}

	AddInternalAPI("apim.v1", Apimv1)

	return true
}

func updateAPIFile() error {
	json := simplejson.New()
	for k, v := range apim.plugins {
		json.Set(k, v)
	}

	f, err := os.OpenFile(apim.apiconf, os.O_TRUNC|os.O_WRONLY, 0644)
	defer f.Close()
	if err != nil {
		return fmt.Errorf("open file %s failed: %v", apim.apiconf, err)
	}

	d, _ := json.MarshalJSON()

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

func getAPI(name string) API {
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
