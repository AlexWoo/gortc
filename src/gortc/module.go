// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//
// Module manager

package main

type Module interface {
	LoadConfig() bool
	Init() bool
	Run()
	State()
	Exit()
}

var modules = make(map[string]Module)
var callseq []string

func AddModule(name string, module Module) {
	modules[name] = module
	callseq = append(callseq, name)
}

func ConfigMdoule() {
	for _, name := range callseq {
		if !modules[name].LoadConfig() {
			LogFatal("Load %s Config failed", name)
		}
	}
}

func InitModule() {
	for _, name := range callseq {
		if !modules[name].Init() {
			LogFatal("Init %s Module failed", name)
		}
	}
}

func RunModule() {
	var main Module
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
