package main

import (
	"context"
	"goldenglow/pkg/database"
	"goldenglow/pkg/messageQueue"
	"goldenglow/pkg/runner"
	"goldenglow/pkg/setup"
	"goldenglow/pkg/userInput"
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
	//Run(
	//	//"archive/start",
	//	//"archive/logic/safe_teach",
	//	//"archive/logic/utils.json",
	//	//"archive/logic/make_attribution/safe_teach/src",
	//	//"archive/logic/make_attribution/safe_teach/test",
	//	//"archive/logic/make_connection/safe_teach/src",
	//	//"archive/logic/make_connection/safe_teach/test",
	//	//"archive/logic/make_question/safe_teach/ask/src",
	//	//"archive/logic/make_question/safe_teach/ask/test",
	//	//"archive/logic/make_question/safe_teach/answer/src",
	//	//"archive/logic/make_question/safe_teach/answer/test",
	//	//"archive/logic/make_answer/safe_teach/ask/src",
	//	//"archive/logic/make_answer/safe_teach/answer/src",
	//	//"archive/logic/make_answer/safe_teach/test",
	//)
	RunWithMsgMgr(
		"archive/logic/make_answer/safe_teach/ask/test",
	)
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
			time.Sleep(100 * time.Millisecond)
		}
	}()
	<-ctx.Done()
	_ = database.DefaultJSONRepo().Shutdown()
	_ = database.DefaultRedisRepo().Shutdown()
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
}
