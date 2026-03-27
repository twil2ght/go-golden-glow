package main

import (
	"bufio"
	"fmt"
	"goldenglow/components/preprocessor"
	"goldenglow/components/queue"
	"goldenglow/components/receiver"
	"goldenglow/components/runner"
	"goldenglow/components/scheduler"
	"goldenglow/components/source"
	"goldenglow/config"
	"goldenglow/node"
	"goldenglow/setup"
	"os"
	"os/signal"
	"syscall"
)

// components
var (
	Queue  = queue.NewQueue()
	Runner = runner.DefaultRunner()
)

// TODO 特殊退出程序的输入
func main() {
	setup.Init()

	exitChan := make(chan os.Signal, 1)
	signal.Notify(exitChan, syscall.SIGINT, syscall.SIGTERM)

	go RunLiteScheduler()

	<-exitChan
	println("\n💤 收到退出信号，开始保存数据...")

	setup.Shutdown()
	println("✅ 数据已保存，程序安全退出")
}

func DefaultSource(sourceReg source.Registry) error {
	keyboardSource := &keyboardInput{
		ch: make(chan string, 10),
	}
	go keyboardSource.readLoop()
	return sourceReg.Register("default", "keyboard", keyboardSource)
}
func DefaultPreprocessor(preprocessReg preprocessor.Registry) error {
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

func RunLiteScheduler() {
	sourceReg := setup.SourceReg
	preprocessReg := setup.PreprocessReg

	err := DefaultPreprocessor(preprocessReg)
	if err != nil {
		panic(err)
	}
	err = DefaultSource(sourceReg)
	if err != nil {
		panic(err)
	}

	mainStream, ok := sourceReg.(source.MainStream)
	if !ok {
		panic("sourceReg does not implement source.MainStream")
	}
	processor, ok := preprocessReg.(preprocessor.Instance)
	if !ok {
		panic("preprocessReg does not implement preprocessor.Instance")
	}

	receiverReg := receiver.NewRegistry(mainStream, sourceReg.Tags())
	defer receiverReg.Shutdown()
	schLite := scheduler.NewLiteScheduler(
		processor,
		Queue,
		Runner,
		node.DefaultFactory())

	rcv, ok := schLite.(receiver.RegisterItem)
	if !ok {
		panic("schLite is not a receiver.RegisterItem")
	}
	err = rcv.OnRegisterReceiver(receiverReg)
	if err != nil {
		panic(err)
	}

	// Start the receiver registry to begin message routing
	receiverReg.Start()

	defer func() {
		err := schLite.Stop()
		if err != nil {
			panic(err)
		}
	}()
	err = schLite.Start()
	if err != nil {
		panic(err)
	}
}
