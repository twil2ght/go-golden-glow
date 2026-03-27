package main

import (
	"bufio"
	"fmt"
	"goldenglow/components/preprocessor"
	"goldenglow/components/queue"
	"goldenglow/components/runner"
	"goldenglow/components/scheduler"
	"goldenglow/components/source"
	"goldenglow/config"
	"goldenglow/dataGen"
	"goldenglow/executor"
	"goldenglow/executor/checker"
	"goldenglow/executor/extractor"
	"goldenglow/lang"
	"goldenglow/node"
	"goldenglow/plugins"
	_ "goldenglow/plugins/builtin/builder"
	_ "goldenglow/plugins/builtin/calculator"
	_ "goldenglow/plugins/builtin/speaker"
	_ "goldenglow/plugins/builtin/timer"
	"goldenglow/storage"
	"os"
)

// registries
var (
	exeReg          = executor.DefaultRegistry()
	exeCheckerReg   = checker.DefaultRegistry()
	exeExtractorReg = extractor.DefaultRegistry()
	langReg         = lang.DefaultRegistry()
	dataGenReg      = dataGen.NewDataGen()
	sourceReg       = source.NewRegistry()
	preprocessReg   = preprocessor.NewRegistry()
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

func DefaultSource() error {
	keyboardSource := &keyboardInput{
		ch: make(chan string),
	}
	go keyboardSource.readLoop()
	return sourceReg.Register("default", "keyboard", keyboardSource)
}
func DefaultPreprocessor() error {
	return preprocessReg.Register("default", "keyboard", func(msg string) string {
		return fmt.Sprintf("%s says %s to %s", config.User, msg, config.GG)
	})
}

type keyboardInput struct {
	ch chan string
}

func (k *keyboardInput) C() <-chan string {
	return k.ch
}

func (k *keyboardInput) readLoop() {
	defer close(k.ch)
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		line := scanner.Text()
		if line == "exit" || line == "quit" {
			break
		}
		k.ch <- line
	}
}
func RunScheduler() {
	err := DefaultPreprocessor()
	if err != nil {
		panic(err)
	}
	err = DefaultSource()
	if err != nil {
		panic(err)
	}

	sched := scheduler.NewScheduler(
		sourceReg.(source.MainStream),
		preprocessReg.(preprocessor.Instance),
		Queue,
		Runner,
		node.DefaultFactory())
	defer func(sched scheduler.Scheduler) {
		err := sched.Stop()
		if err != nil {
			panic(err)
		}
	}(sched)
	err = sched.Start()
	if err != nil {
		panic(err)
	}
}
