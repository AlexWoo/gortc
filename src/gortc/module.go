// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//
// Module manager

package main

import (
	"fmt"
	"runtime"
)

type Module interface {
	LoadConfig() bool
	Init() bool
	Run()
	State()
	Exit()
}

var modules = make(map[string]Module)

func AddModule(name string, module Module) {
	modules[name] = module
}

func ConfigMdoule() {
	for name, module := range modules {
		if !module.LoadConfig() {
			// TODO logErr
			fmt.Printf("Load %s Config failed", name)
			runtime.Goexit()
		}
	}
}

func InitModule() {
	for name, module := range modules {
		if !module.Init() {
			//TODO logErr
			fmt.Printf("Init %s Module failed", name)
			runtime.Goexit()
		}
	}
}

func RunModule() {
	var main *MainModule
	for name, module := range modules {
		if name == "main" {
			// Make sure main module run at last
			main = module
			continue
		}

		// other module should run in goroutine
		module.Run()
	}

	main.Run()
}

func StateModule(name string) {
	if name == "" {
		for _, module := range modules {
			module.State()
		}
		return
	}

	module, ok := modules[name]
	if !ok {
		return
	}

	module.State()
}

func ExitModule() {
	for _, module := range modules {
		module.Exit()
	}
}
