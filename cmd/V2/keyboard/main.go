package main

import (
	"context"
	"fmt"
	"goldenglow/config"
	"goldenglow/pkg/database"
	"goldenglow/pkg/messageQueue"
	"goldenglow/pkg/node/template"
	"goldenglow/pkg/runner"
	"goldenglow/pkg/setup"
	"goldenglow/pkg/tui"
	"goldenglow/plugin"
	"goldenglow/utils"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

var (
	cacheLogPath = filepath.Join(utils.RootDir, "dialogue_history.log")
	workNum      = 1
)

func main() {
	bg := setup.Init()
	msgQueue := messageQueue.New(cacheLogPath)
	consumer := runner.New(workNum, msgQueue, bg.NodeFactory)
	chatUI := tui.NewChat()
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Display system messages coming through the MsgQueueMgr.
	// Parse speaker-formatted messages to strip redundant prefixes like "Susie says to Zero :"
	bg.MsgQueueMgr.OnMessage(func(rawMsg string) {
		var (
			sender = "Background"
			msg    = rawMsg
		)
		// "Susie says to Zero : hello"
		if ok, varSet := template.MatchTemplate(rawMsg, "$1 says to $2 : $3"); ok {
			sender, msg = varSet["$1"].Value(), varSet["$3"].Value()
		} else if ok, varSet := template.MatchTemplate(rawMsg, "$1 say : $2"); ok {
			// "Susie says : hello"
			sender = varSet["$1"].Value()
			if sender != config.GG {
				msg = fmt.Sprintf("[Raw] %s", msg)
			}
		}
		if sender != config.GG && sender != config.User {
			return
		}
		chatUI.Display(tui.Message{Sender: sender, Text: msg, Time: time.Now()})
	})

	go bg.MsgQueueMgr.Start(msgQueue, ctx)
	go consumer.Run(ctx)

	// Route TUI input to the message queue
	go func() {
		for input := range chatUI.InputCh() {
			msgQueue.Add(WithUserPrefix(config.User, input))
		}
	}()

	// Run the TUI (blocks until ctx is done)
	chatUI.Start(ctx)

	_ = database.DefaultJSONRepo().Shutdown()
	_ = database.DefaultRedisRepo().Shutdown()
	plugin.DefaultManager.Range(func(_ string, p plugin.Interface) bool {
		p.Shutdown()
		return true
	})
}
func WithUserPrefix(user, msg string) string {
	return fmt.Sprintf("[Raw] %s say : %s", user, msg)
}
