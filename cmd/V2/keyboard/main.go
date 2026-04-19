package main

import (
	"context"
	"goldenglow/pkg/messageQueue"
	"goldenglow/pkg/runner"
	"goldenglow/pkg/setup"
	"goldenglow/pkg/userInput"
	"goldenglow/utils"
	"path/filepath"
)

var (
	cacheLogPath = filepath.Join(utils.RootDir, "dialogue_history.log")
	workNum      = 5
)
var (
	bg          = setup.Init()
	msgQueueMgr = bg.MsgQueueMgr
	msgQueue    = messageQueue.New(cacheLogPath)
	consumer    = runner.New(workNum, msgQueue, bg.NodeFactory)
	keyboard    = userInput.NewKeyboard(msgQueue)
	ctx         = context.Background()
)

func main() {
	go msgQueueMgr.Start(msgQueue, ctx)
	go consumer.Run(ctx)
	go keyboard.Start(ctx)
	<-ctx.Done()
}
