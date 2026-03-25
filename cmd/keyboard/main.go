package main

import (
	"fmt"
	"goldenglow/components/preprocessor"
	"goldenglow/components/queue"
	"goldenglow/components/runner"
	"goldenglow/components/source"
	"goldenglow/dataGen"
	"goldenglow/executor"
	"goldenglow/lang"
	"goldenglow/plugins"
	_ "goldenglow/plugins/builtin/builder"
	_ "goldenglow/plugins/builtin/speaker"
	"goldenglow/storage"
)

// registries
// TODO add logger to all registries
var (
	exeReg        = executor.DefaultRegistry()
	langReg       = lang.DefaultRegistry()
	dataGenReg    = dataGen.NewDataGen()
	sourceReg     = source.NewRegistry()
	preprocessReg = preprocessor.NewRegistry()
)

// components
var (
	Queue  = queue.NewQueue()
	Runner = runner.DefaultRunner()
)

func main() {
	defer Shutdown()
	Init()
}

func Init() {
	if err := storage.DefaultJSONRepo().Init(); err != nil {
		panic(err)
	}
	plugins.Init()
	pluginInstances := plugins.GetAll()
	for _, pluginInstance := range pluginInstances {
		parsePlugin(pluginInstance)
	}
	if err := Run(); err != nil {
		panic(err)
	}
}
func parsePlugin(plugin plugins.Item) {
	var (
		pluginName = plugin.Name()
	)
	if err := plugin.OnRegisterDataGen(dataGenReg); err != nil {
		panic(fmt.Sprintf("%s.OnRegisterDataGen err:%v", pluginName, err))
	}
	if err := plugin.OnRegisterLang(langReg); err != nil {
		panic(fmt.Sprintf("%s.OnRegisterLang err:%v", pluginName, err))
	}
	if err := plugin.OnRegisterExecutor(exeReg); err != nil {
		panic(fmt.Sprintf("%s.OnRegisterExecutor err:%v", pluginName, err))
	}
	if err := plugin.OnRegisterPreprocessor(preprocessReg); err != nil {
		panic(fmt.Sprintf("%s.OnRegisterPreprocessor err:%v", pluginName, err))
	}
	if err := plugin.OnRegisterInputSource(sourceReg); err != nil {
		panic(fmt.Sprintf("%s.OnRegisterInputSource err:%v", pluginName, err))
	}
}
func Run() error {
	if err := exeReg.RunAll(); err != nil {
		return err
	}
	if err := dataGenReg.RunAll(); err != nil {
		return err
	}
	if err := langReg.RunAll(); err != nil {
		return err
	}
	return nil
}
func Shutdown() {
	err := storage.DefaultJSONRepo().Shutdown()
	if err != nil {
		panic(err)
	}
}
