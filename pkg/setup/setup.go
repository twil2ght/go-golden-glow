package setup

import (
	"goldenglow/pkg/messageQueue"
	"goldenglow/pkg/node"
	"goldenglow/pkg/node/handler"
	"goldenglow/plugin"
	_ "goldenglow/plugin/mount"
)

type Background struct {
	MsgQueueMgr messageQueue.Manager
	NodeFactory node.Factory
}

func Init() *Background {
	var (
		pluginMgr   = plugin.DefaultManager
		executor    = handler.NewExecutor()
		checker     = handler.NewChecker()
		extractor   = handler.NewExtractor()
		msgQueueMgr = messageQueue.NewManager()
		nodeFactory = node.DefaultFactory
	)
	pluginMgr.Range(func(key string, item plugin.Interface) bool {
		if e, ok := item.(handler.ExecuteHook); ok {
			e.OnRegisterExecutor(executor)
		}
		if e, ok := item.(handler.CheckHook); ok {
			e.OnCheckExecutor(checker)
		}
		if e, ok := item.(handler.ExtractHook); ok {
			e.OnExtractExecutor(extractor)
		}
		if e, ok := item.(messageQueue.MsgQueueHook); ok {
			e.OnRegisterMsgProvider(msgQueueMgr)
		}
		item.Init()
		return true
	})
	executor.OnRegisterFactory(nodeFactory)
	checker.OnRegisterFactory(nodeFactory)
	extractor.OnRegisterFactory(nodeFactory)
	return &Background{
		MsgQueueMgr: msgQueueMgr,
		NodeFactory: nodeFactory,
	}
}
