package messageQueue

import (
	"goldenglow/pkg/log"
	"goldenglow/pkg/workqueue"
)

var (
	logger = log.Default()
)

type Interface interface {
	workqueue.Interface[string]
}
type msgQueue struct {
	workqueue.Interface[string]
	logs      []string
	cachePath string
}

func (mq *msgQueue) Add(value string) {
	logger.Debug("msgQueue Add() called", "module", "msgQueue", "value", value)
	mq.Interface.Add(value)
	mq.addToLog(value)
}
func (mq *msgQueue) Get() (item string, shutdown bool) {
	item, shutdown = mq.Interface.Get()
	logger.Debug("msgQueue Get() called", "module", "msgQueue", "value", item)
	return item, shutdown
}
func (mq *msgQueue) addToLog(value string) {
	mq.logs = append(mq.logs, value)
}
func (mq *msgQueue) Save() {
	//TODO
}
func (mq *msgQueue) Shutdown() {
	mq.Interface.Shutdown()
	mq.Save()
}
func New(cachePath string) Interface {
	return &msgQueue{
		Interface: workqueue.New[string](),
		cachePath: cachePath,
	}
}
