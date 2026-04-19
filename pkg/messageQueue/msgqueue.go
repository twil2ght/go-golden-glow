package messageQueue

import (
	"goldenglow/pkg/workqueue"
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
	mq.Interface.Add(value)
	mq.addToLog(value)
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
