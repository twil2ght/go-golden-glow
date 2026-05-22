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
	//"src/pkg/Executor/utils.json",
	//"src/pkg/Executor/util_test.json",
	//"src/Grammar/template_test.json",
	//"src/pkg/Cond/utils.json",
	//"src/pkg/Cond/utils_test.json",
	//"src/Grammar/Object_repo.json",
	//"src/Grammar/Verb_Phrase.json",
	//"src/Grammar/Verb_Phrase_test.json",
	//"src/interaction/Cond.json",
	//"src/test/2.json",
	//"src/interaction/Res.json",
	//"src/Grammar/Object_IO+DO_to.json",
	//"src/test/1.json",
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
