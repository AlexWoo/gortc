// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//
// Module manager

package rtclib

type Module interface {
	LoadConfig(rtcPath string) bool
	Init() bool
	Run()
	Exit()
}

var modules = make(map[string]Module)
var callseq []string

func AddModule(name string, module Module) {
	modules[name] = module
	callseq = append(callseq, name)
}

func InitModule(log *Log, rtcPath string) {
	for _, name := range callseq {
		if !modules[name].LoadConfig(rtcPath) {
			log.LogFatal("Module %s load config failed", name)
		}

		if !modules[name].Init() {
			log.LogFatal("Module %s init failed", name)
		}

		log.LogInfo("Module %s init successd", name)
	}
}

func RunModule(log *Log) {
	for name, module := range modules {
		go func() {
			module.Run()
			log.LogInfo("Module %s gracefully exit", name)
			delete(modules, name)
		}()
	}
}

func CheckModule() bool {
	return len(modules) != 0
}

func ExitModule() {
	for _, module := range modules {
		module.Exit()
	}
}
