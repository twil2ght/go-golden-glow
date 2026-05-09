package main

import (
	"context"
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
		//"archive/logic/engines/UserInterface/SafeCalculate.json",
		//"archive/logic/engines/UserInterface/SafeCalculate_test.json",
		//"archive/logic/engines/UserInterface/SetByStep.json",
		//"archive/logic/engines/UserInterface/SetByStep_test.json",
		//"archive/logic/engines/UserInterface/CondGroup_Iterator.json",
		"archive/logic/engines/UserInterface/CondGroup_Iterator_test.json",
	)
	//RunWithMsgMgr(
	//
	//)
}
func Run(dataDir ...string) {
	go consumer.Run(ctx)
	go func() {
		for _, dir := range dataDir {
			file.Run(dir)
		}
		for {
			if msgQueue.Len() == 0 {
				cancel()
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
	time.Sleep(1 * time.Second)
	cancel()
	_ = database.DefaultJSONRepo().Shutdown()
	_ = database.DefaultRedisRepo().Shutdown()
	plugin.DefaultManager.Range(func(_ string, p plugin.Interface) bool {
		p.Shutdown()
		return true
	})
}
