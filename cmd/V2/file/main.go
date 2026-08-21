package main

import (
	"context"
	"fmt"
	"goldenglow/pkg/database"
	"goldenglow/pkg/messageQueue"
	"goldenglow/pkg/runner"
	"goldenglow/pkg/setup"
	"goldenglow/pkg/userInput"
	"goldenglow/plugin"
	"time"
)

var (
	cacheLogPath = "dialogue_history.log"
	workNum      = 5
)
var (
	bg          = setup.Init()
	msgQueue    = messageQueue.New(cacheLogPath)
	consumer    = runner.New(workNum, msgQueue, bg.NodeFactory)
	file        = userInput.NewFile(msgQueue)
	ctx, cancel = context.WithCancel(context.Background())
)

func main() {
	Run(
		//"src/Global.json",
		//"src/BetterSettingWords/const.json",
		"src/test/new/0_learn_word.json",
	)
	//RunWithMsgMgr(
	//	"src/test/5_test_TODO",
	//)
}
func Run(dataDir ...string) {
	go consumer.Run(ctx)
	go func() {
		file.Count = 0
		for _, dir := range dataDir {
			file.Run(dir)
		}
		for {
			if msgQueue.Len() == 0 && consumer.IsFinished() {
				cancel()
				fmt.Printf("Commands Loaded: %d\n", file.Count)
				return
			}
			time.Sleep(1000 * time.Millisecond)
		}
	}()
	<-ctx.Done()
	_ = database.DefaultJSONRepo().Shutdown()
	_ = database.DefaultRedisRepo().Shutdown()
	plugin.DefaultManager.Range(func(_ string, p plugin.Interface) bool {
		p.Shutdown()
		return true
	})
}
func RunWithMsgMgr(dataDir ...string) {
	go bg.MsgQueueMgr.Start(msgQueue, ctx)
	go consumer.Run(ctx)
	go func() {
		for _, dir := range dataDir {
			file.Run(dir)
		}
	}()
	time.Sleep(2 * time.Second)
	cancel()
	_ = database.DefaultJSONRepo().Shutdown()
	_ = database.DefaultRedisRepo().Shutdown()
	plugin.DefaultManager.Range(func(_ string, p plugin.Interface) bool {
		p.Shutdown()
		return true
	})
}
