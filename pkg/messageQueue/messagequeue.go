package messageQueue

import (
	"goldenglow/pkg/workqueue"
)

type Interface interface {
	workqueue.Interface[string]
}
type messageQueue struct {
	workqueue.Interface[string]
	logs      []string
	cachePath string
}

func (mq *messageQueue) Add(value string) {
	mq.Interface.Add(value)
	mq.addToLog(value)
}
func (mq *messageQueue) addToLog(value string) {
	mq.logs = append(mq.logs, value)
}
func (mq *messageQueue) Save() {

}
func (mq *messageQueue) Shutdown() {
	mq.Interface.Shutdown()
	mq.Save()
}
func New() Interface {
	return &messageQueue{
		Interface: workqueue.New[string](),
	}
}
