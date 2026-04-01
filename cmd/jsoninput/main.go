package main

import (
	"encoding/json"
	"fmt"
	"goldenglow/components/collector"
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
var clt = collector.New()

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: jsoninput <path-to-json-file>")
		fmt.Println("JSON file format: {\"inputs\": [\"message1\", \"message2\", ...]}")
		os.Exit(1)
	}

	jsonPath := os.Args[1]

	setup.Init()

	exitChan := make(chan os.Signal, 1)
	signal.Notify(exitChan, syscall.SIGINT, syscall.SIGTERM)

	go RunLiteScheduler(jsonPath)
	clt.SetSource("default:jsoninput", "speaker:self")
	go clt.Run()
	<-exitChan
	err := clt.Save()
	if err != nil {
		panic(err)
	}
	println("\n💤 saving data...")

	setup.Shutdown()
	println("✅ shutdown complete")
}

func DefaultSource(sourceReg source.Registry, jsonPath string) error {
	jsonSource := &jsonInput{
		ch: make(chan string, 10),
	}
	go jsonSource.readFromFile(jsonPath)
	return sourceReg.Register("default", "jsoninput", jsonSource)
}

func DefaultPreprocessor(preprocessReg preprocessor.Registry) error {
	return preprocessReg.Register("default", "jsoninput", func(msg string) string {
		return fmt.Sprintf("%s says %s to %s", config.User, msg, config.GG)
	})
}

type jsonInput struct {
	ch chan string
}

func (j *jsonInput) C() <-chan string {
	return j.ch
}

func (j *jsonInput) readFromFile(jsonPath string) {
	defer close(j.ch)

	data, err := os.ReadFile(jsonPath)
	if err != nil {
		fmt.Printf("Error reading JSON file: %v\n", err)
		return
	}

	var jsonData struct {
		Inputs []string `json:"inputs"`
	}

	err = json.Unmarshal(data, &jsonData)
	if err != nil {
		fmt.Printf("Error parsing JSON file: %v\n", err)
		return
	}

	for _, input := range jsonData.Inputs {
		if input == "exit" || input == "quit" {
			break
		}
		j.ch <- input
	}
}

func RunLiteScheduler(jsonPath string) {
	sourceReg := setup.SourceReg
	preprocessReg := setup.PreprocessReg

	err := DefaultPreprocessor(preprocessReg)
	if err != nil {
		panic(err)
	}
	err = DefaultSource(sourceReg, jsonPath)
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

	collectRcv, ok := clt.(receiver.RegisterItem)
	if !ok {
		panic("collector is not a receiver.RegisterItem")
	}
	err = collectRcv.OnRegisterReceiver(receiverReg)
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

	// Block forever to keep the scheduler running
	// The defer statements will run when the program exits
	select {}
}
