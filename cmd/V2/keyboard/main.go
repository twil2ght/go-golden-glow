package main

import (
	"context"
	"goldenglow/config"
	"goldenglow/pkg/log"
	"goldenglow/pkg/messageQueue"
	"goldenglow/pkg/runner"
	"goldenglow/pkg/setup"
	"goldenglow/pkg/tui"
	"goldenglow/utils"
	"log/slog"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

var (
	cacheLogPath = filepath.Join(utils.RootDir, "dialogue_history.log")
	workNum      = 5
)

func main() {
	bg := setup.Init()
	msgQueue := messageQueue.New(cacheLogPath)
	consumer := runner.New(workNum, msgQueue, bg.NodeFactory)
	chatUI := tui.NewChat()
	log.SetLevel(slog.LevelError)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Display system messages coming through the MsgQueueMgr
	bg.MsgQueueMgr.OnMessage(func(msg string) {
		chatUI.Display(tui.Message{Sender: config.GG, Text: msg, Time: time.Now()})
	})

	go bg.MsgQueueMgr.Start(msgQueue, ctx)
	go consumer.Run(ctx)

	// Route TUI input to the message queue
	go func() {
		for input := range chatUI.InputCh() {
			msgQueue.Add(input)
		}
	}()

	// Run the TUI (blocks until ctx is done)
	chatUI.Start(ctx)
}
