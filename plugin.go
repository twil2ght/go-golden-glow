package main

import (
	"goldenglow/plugins"
	betterspeak "goldenglow/plugins/betterSpeak"
	"goldenglow/plugins/checker"
	"goldenglow/plugins/dictionary"
	"goldenglow/plugins/kv"
	"goldenglow/plugins/pack"
	"goldenglow/plugins/timer"
)

func PluginRegister() {
	// better checker plugin
	myPlugins.Register(func(ctx *plugins.Context) {
		var (
			CHECKS = checker.NewRegistry()
		)
		err := checker.NewChecker(CHECKS).Register(ctx)
		if err != nil {
			return
		}

		CHECKS.Register(
			"timer_with_in_range",
			timer.WithinRange,
		)
	})

	// better speaker plugin
	myPlugins.Register(func(ctx *plugins.Context) {
		var (
			newSpeaker = betterspeak.NewEngine(nil)
		)
		scheduler.Collector(newSpeaker)
		err := betterspeak.New(newSpeaker).Register(ctx)
		if err != nil {
			return
		}

	})

	// dictionary plugin
	myPlugins.Register(func(ctx *plugins.Context) {
		err := dictionary.New(dictionary.NewEngine(dictionary.FilePath)).Register(ctx)
		if err != nil {
			return
		}
	})
	myPlugins.RegisterHook(dictionary.ShutDown)

	// pack plugin
	myPlugins.Register(func(ctx *plugins.Context) {
		err := pack.New(pack.NewEngine(nil)).Register(ctx)
		if err != nil {
			return
		}
	})

	// timer plugin
	myPlugins.Register(func(ctx *plugins.Context) {
		err := timer.New(timer.NewEngine()).Register(ctx)
		if err != nil {
			return
		}
	})

	//KV plugin
	myPlugins.Register(func(ctx *plugins.Context) {
		err := kv.New(&kv.Engine{}).Register(ctx)
		if err != nil {
			return
		}
	})
}
