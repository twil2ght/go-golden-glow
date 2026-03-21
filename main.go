package main

import (
	"bufio"
	"fmt"
	"goldenglow/data"
	"goldenglow/node"
	"goldenglow/pinkcat"
	"goldenglow/plugin"
	"goldenglow/plugins"
	betterspeak "goldenglow/plugins/betterSpeak"
	"goldenglow/storage"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

var Demo = []string{
	"if you don't know what something is then you should say what is it to me",
	// "if you say what is something to me then you ask me about what it is",
}
var Inputs = []string{
	"if you say what is something to me then you ask me a question",
	"if you say what is something to me then your question is what something is",
	// "if I say something is A then i answer your question",
	"if I say something is A and your question is what something is then i answer your question",
	// "if you ask me a question then wait until I answer your question",
	"if you ask me a question then you should wait", //user go on;output shutdown
	"if I answer your question then you should stop wait",
	"[Pack_In] if you ask me a question then you should wait",
	"[Pack_In] if I answer your question then you should stop wait",
	"[Pack_Close] if you ask me a question then you should wait util i answer your question", //[P] Pack_In... [R] | Pack_Close
	"[BackPack_In] I say something is A",
	"[BackPack_In] your question is what something is",
	"[BackPack_close] i answer your question", //[P] [BackPack_Close] [R] | [PackPack_In]...
}
var (
	user    = 0
	scanner = bufio.NewScanner(os.Stdin)
	ticker  = time.NewTicker(60 * time.Second)
	engine  = pinkcat.Single(user)

	scheduler = betterspeak.NewScheduler(engine, user)
	myPlugins = plugins.NewRegistry()
)
var (
	JSONRepo     = storage.NewJSONRepo("", "")
	JSONLiteRepo = JSONRepo.(storage.LightRepository)
)

func main() {
	data.InitMarks()
	defer ticker.Stop()
	defer myPlugins.ShutDown()

	PluginRegister()
	myPlugins.Init(
		plugins.WithRunner(engine),
		plugins.WithUserID(user),
		plugins.WithLangRegistry(plugin.NewLangRegistry(nil)),
		plugins.WithNodeRegistry(node.NewFactory(nil, nil)),
	)

	go MainProcess()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	for {
		fmt.Print("-->")
		if !scanner.Scan() {
			break
		}
		resp := strings.TrimSpace(scanner.Text())
		if resp != "" {
			scheduler.Set(resp)
		} else {
			break
		}
	}
}

func MainProcess() {
	for {
		time.Sleep(1 * time.Second)
		Input := scheduler.Exhaust()
		if Input != nil {
			if Input.Dec == nil {
				//user
				err := JSONLiteRepo.Set("[speaker]", fmt.Sprintf("%d", user))
				if err != nil {
					return
				}
				err = JSONLiteRepo.Set("[listener]", "[GG]")
				if err != nil {
					return
				}
				err = scheduler.Runner.Run()
				if err != nil {
					return
				}

				err = JSONLiteRepo.Set("[listener]", fmt.Sprintf("%d", user))
				if err != nil {
					return
				}
				err = JSONLiteRepo.Set("[speaker]", "[GG]")
				if err != nil {
					return
				}
				runner := scheduler.Runner
				prevDecorator := runner.GetDecorator()
				runner.SetDecorator(func(s string) string { return "[I Ask] " + s })
				// TODO [ASK]
				// for t := range strings.FieldsSeq(Input.Val) {
				// 	if _, ok := dictionary.GlobalDictionary.Table[t]; ok {
				// 		continue
				// 	}
				// 	// fmt.Printf("\n[SLN]:%s\n-->", t)
				// 	engine.Run(t)
				// }
				runner.SetDecorator(prevDecorator)
			} else {
				//[GG]
				fmt.Printf("\n[Speaker]:%s\n-->", Input.Val)
				err := JSONLiteRepo.Set("[listener]", fmt.Sprintf("%d", user))
				if err != nil {
					return
				}
				err = JSONLiteRepo.Set("[speaker]", "[GG]")
				if err != nil {
					return
				}
				prevDecorator := scheduler.Runner.GetDecorator()
				scheduler.Runner.SetDecorator(Input.Dec)
				err = scheduler.Runner.Run()
				if err != nil {
					return
				}
				scheduler.Runner.SetDecorator(prevDecorator)
			}
		}
	}
}
