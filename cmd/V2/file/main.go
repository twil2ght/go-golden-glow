package main

import (
	"context"
	"goldenglow/pkg/database"
	"goldenglow/pkg/messageQueue"
	"goldenglow/pkg/runner"
	"goldenglow/pkg/setup"
	"goldenglow/pkg/userInput"
	"goldenglow/utils"
	"path/filepath"
	"time"
)

var (
	cacheLogPath = RelPath("dialogue_history.log")
	workNum      = 5
	dataDir      = RelPath("archive/logic/make_connection/safe_teach/test")
)
var (
	bg          = setup.Init()
	msgQueue    = messageQueue.New(cacheLogPath)
	consumer    = runner.New(workNum, msgQueue, bg.NodeFactory)
	file        = userInput.NewFile(msgQueue)
	ctx, cancel = context.WithCancel(context.Background())
)

func main() {
	//Init()
	Run(
		//"archive/logic/make_connection/safe_teach/src",
		//"archive/logic/make_question/safe_teach/ask/src",
		"archive/logic/make_question/safe_teach/ask/test",
	)
}
func Init() {
	Run(
		RelPath("archive/start"),
		RelPath("archive/logic/safe_teach"),
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
func RelPath(path string) string {
	return filepath.Join(utils.RootDir, path)
}
