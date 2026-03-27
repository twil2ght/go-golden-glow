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
	"goldenglow/node"
	_ "goldenglow/plugins/builtin/builder"
	_ "goldenglow/plugins/builtin/calculator"
	_ "goldenglow/plugins/builtin/speaker"
	_ "goldenglow/plugins/builtin/timer"
	"os"
)

// registries
var (
	sourceReg     = source.NewRegistry()
	preprocessReg = preprocessor.NewRegistry()
)

// components
var (
	Queue  = queue.NewQueue()
	Runner = runner.DefaultRunner()
)

func main() {

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
