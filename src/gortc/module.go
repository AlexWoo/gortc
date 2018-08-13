// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//
// Module manager

package main

import "rtclib"

var modules = make(map[string]rtclib.Module)
var callseq []string

func addModule(name string, module rtclib.Module) {
	modules[name] = module
	callseq = append(callseq, name)
}

func initModule() {
	for _, name := range callseq {
		if !modules[name].LoadConfig() {
			LogFatal("Module %s load config failed", name)
		}

		if !modules[name].Init(mainlogCtx.log) {
			LogFatal("Module %s init failed", name)
		}

		LogInfo("Module %s init successd", name)
	}
}

func runModule() {
	for name, module := range modules {
		go func(name string, module rtclib.Module) {
			module.Run()
			LogInfo("Module %s gracefully exit", name)
			delete(modules, name)
		}(name, module)
	}
}

func checkModule() bool {
	return len(modules) != 0
}

func exitModule() {
	for _, module := range modules {
		module.Exit()
	}
}
