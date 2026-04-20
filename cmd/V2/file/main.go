package main

import (
	"context"
	"goldenglow/pkg/messageQueue"
	"goldenglow/pkg/runner"
	"goldenglow/pkg/setup"
	"goldenglow/pkg/userInput"
	"goldenglow/storage"
	"goldenglow/utils"
	"path/filepath"
	"time"
)

var (
	cacheLogPath = filepath.Join(utils.RootDir, "dialogue_history.log")
	workNum      = 5
	dataDir      = filepath.Join(utils.RootDir, "archive/info/grammar/make_attribution/safe_teach")
)
var (
	bg          = setup.Init()
	msgQueue    = messageQueue.New(cacheLogPath)
	consumer    = runner.New(workNum, msgQueue, bg.NodeFactory)
	file        = userInput.NewFile(msgQueue)
	ctx, cancel = context.WithCancel(context.Background())
)

func main() {
	go consumer.Run(ctx)
	go func() {
		file.Run(dataDir)
		for {
			if msgQueue.Len() == 0 {
				cancel()
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()
	<-ctx.Done()
	_ = storage.DefaultJSONRepo().Shutdown()
	_ = storage.DefaultRedisRepo().Shutdown()
}
