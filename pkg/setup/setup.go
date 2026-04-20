package setup

import (
	"goldenglow/pkg/datagen"
	"goldenglow/pkg/dataloader"
	"goldenglow/pkg/messageQueue"
	"goldenglow/pkg/node"
	"goldenglow/pkg/node/handler"
	"goldenglow/pkg/node/template"
	"goldenglow/plugin"
	_ "goldenglow/plugin/mount"
	"goldenglow/storage"
)

type Background struct {
	MsgQueueMgr messageQueue.Manager
	NodeFactory node.Factory
}

func Init() *Background {
	_ = storage.DefaultJSONRepo().Init()
	_ = storage.DefaultRedisRepo().Init()
	var (
		dataDir     = datagen.RootDir
		pluginMgr   = plugin.DefaultManager
		executor    = handler.NewExecutor()
		checker     = handler.NewChecker()
		extractor   = handler.NewExtractor()
		msgQueueMgr = messageQueue.NewManager()
		dataGen     = datagen.NewGenerator()
		dataLoader  = dataloader.Default()
		nodeFactory = node.DefaultFactory
		conflictMgr = template.DefaultConflictManager
	)
	pluginMgr.Range(func(key string, item plugin.Interface) bool {
		if e, ok := item.(handler.ExecuteHook); ok {
			e.OnRegisterExecutor(executor)
		}
		if e, ok := item.(handler.CheckHook); ok {
			e.OnRegisterChecker(checker)
		}
		if e, ok := item.(handler.ExtractHook); ok {
			e.OnRegisterExtractor(extractor)
		}
		if e, ok := item.(messageQueue.MsgQueueHook); ok {
			e.OnRegisterMsgProvider(msgQueueMgr)
		}
		if e, ok := item.(datagen.Hook); ok {
			e.OnRegisterDataGen(dataGen)
		}
		if e, ok := item.(template.Hook); ok {
			e.OnRegisterConflictRule(conflictMgr)
		}
		item.Init()
		return true
	})
	executor.OnRegisterFactory(nodeFactory)
	checker.OnRegisterFactory(nodeFactory)
	extractor.OnRegisterFactory(nodeFactory)
	dataGen.Run()
	dataLoader.Load(dataDir)
	return &Background{
		MsgQueueMgr: msgQueueMgr,
		NodeFactory: nodeFactory,
	}
}
