package setup

import (
	"fmt"
	"goldenglow/components/preprocessor"
	"goldenglow/components/source"
	"goldenglow/dataGen"
	"goldenglow/executor"
	"goldenglow/executor/checker"
	"goldenglow/executor/extractor"
	"goldenglow/lang"
	"goldenglow/plugins"
	_ "goldenglow/plugins/builtin/builder"
	_ "goldenglow/plugins/builtin/calculator"
	_ "goldenglow/plugins/builtin/speaker"
	_ "goldenglow/plugins/builtin/timer"
	"goldenglow/storage"
)

var (
	exeReg          = executor.DefaultRegistry()
	exeCheckerReg   = checker.DefaultRegistry()
	exeExtractorReg = extractor.DefaultRegistry()
	langReg         = lang.DefaultRegistry()
	dataGenReg      = dataGen.NewDataGen()
	sourceReg       = source.NewRegistry()
	preprocessReg   = preprocessor.NewRegistry()
)

func Init() {
	if err := storage.DefaultJSONRepo().Init(); err != nil {
		panic(err)
	}
	plugins.Init()
	pluginInstances := plugins.GetAll()
	for _, pluginInstance := range pluginInstances {
		parsePlugin(pluginInstance)
	}
	if err := run(); err != nil {
		panic(err)
	}
}
func parsePlugin(plugin plugins.Item) {
	pluginName := plugin.Name()

	mustRegs := []struct {
		name string
		fn   func() error
	}{
		{"OnRegisterDataGen", func() error { return plugin.OnRegisterDataGen(dataGenReg) }},
		{"OnRegisterLang", func() error { return plugin.OnRegisterLang(langReg) }},
		{"OnRegisterExecutor", func() error { return plugin.OnRegisterExecutor(exeReg) }},
		{"OnRegisterPreprocessor", func() error { return plugin.OnRegisterPreprocessor(preprocessReg) }},
		{"OnRegisterInputSource", func() error { return plugin.OnRegisterInputSource(sourceReg) }},
	}

	for _, reg := range mustRegs {
		if err := reg.fn(); err != nil {
			panic(fmt.Sprintf("[%s] %s failed: %v", pluginName, reg.name, err))
		}
	}
	if checkerPlugin, ok := plugin.(checker.RegisterItem); ok {
		if err := checkerPlugin.OnRegisterChecker(exeCheckerReg); err != nil {
			panic(fmt.Sprintf("[%s] OnRegisterChecker failed: %v", pluginName, err))
		}
	}
	if extractorPlugin, ok := plugin.(extractor.RegisterItem); ok {
		if err := extractorPlugin.OnRegisterExtractor(exeExtractorReg); err != nil {
			panic(fmt.Sprintf("[%s] OnRegisterExtractor failed: %v", pluginName, err))
		}
	}
}
func run() error {
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
