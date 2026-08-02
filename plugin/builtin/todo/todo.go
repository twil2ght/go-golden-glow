// Package todo
//
//	Items that added during runtime won't be saved if they are consumed before the end
//	Thus,those have not been consumed remain in file
package todo

import (
	"encoding/json"
	"fmt"
	"goldenglow/pkg/datagen"
	"goldenglow/pkg/log"
	"goldenglow/pkg/messageQueue"
	"goldenglow/pkg/node/handler"
	"goldenglow/pkg/workqueue"
	"goldenglow/plugin"
	"goldenglow/utils"
	"os"
	"path/filepath"
	"time"
)

func init() {
	plugin.DefaultManager.Register(name, New())
}

const (
	name = "TODO"

	keyAction = "action"
	keyValue  = "value"
	keyCaller = "caller"

	actionAdd               = "Add"
	actionAddWithPostCaller = "AddWithPostCaller"

	msgChannelBuffer = 1000
	tickInterval     = 20 * time.Millisecond
)

type todo struct {
	workQueue  workqueue.Interface[string]
	msgChannel chan string
	shutdown   chan struct{}
}

func (t *todo) Init() {
	t.loadFromDisk()
	if t.workQueue.Len() > 0 {
		fmt.Printf("Start TODO Queue\n")
		go t.tickLoop()
	}
}

func (t *todo) Shutdown() {
	close(t.shutdown)
	t.saveToDisk()
}

func (t *todo) OnRegisterExecutor(reg handler.Executor[handler.ExecuteHandler]) {
	reg.Handlers().Register(name, func(parameters handler.Parameters) {
		action, _ := parameters.Get(keyAction)
		switch action {
		case actionAdd:
			{
				value, _ := parameters.Get(keyValue)
				caller, _ := parameters.Get(keyCaller)
				if value == "" {
					return
				}
				raw := fmt.Sprintf("%s @Caller %s", value, caller)
				t.workQueue.Add(raw)
			}
		case actionAddWithPostCaller:
			{
				value, _ := parameters.Get(keyValue)
				caller, _ := parameters.Get(keyCaller)
				if value == "" {
					return
				}
				raw := fmt.Sprintf("%s @ %s", value, caller)
				t.workQueue.Add(raw)
			}
		}
	})
}

func (t *todo) OnRegisterMsgProvider(reg messageQueue.Manager) {
	reg.Add(name, t.msgChannel)
}

func (t *todo) OnRegisterDataGen(gen datagen.Generator) {
	provider := datagen.NewProvider()
	provider.Add("add", datagen.NewData(
		[]string{"[TODO:Add] $1 @Caller $2"},
		[]string{},
		map[string]string{
			keyAction: actionAdd,
			keyValue:  "$1",
			keyCaller: "$2",
		},
		datagen.AsExecutor,
	))
	provider.Add("addWith@", datagen.NewData(
		[]string{"[TODO:Add] $1 @ $2"},
		[]string{},
		map[string]string{
			keyAction: actionAddWithPostCaller,
			keyValue:  "$1",
			keyCaller: "$2",
		},
		datagen.AsExecutor,
	))
	gen.AddProvider(name, provider)
}

func (t *todo) tickLoop() {
	for {
		select {
		case <-t.shutdown:
			return
		default:
		}
		item, shutdown := t.workQueue.Get()
		if shutdown {
			return
		}
		select {
		case t.msgChannel <- item:
			fmt.Printf("TODO:Route %s to MsgCH\n", item)
		case <-t.shutdown:
			return
		}
		time.Sleep(tickInterval)
	}
}

var persistDir = filepath.Join(utils.RootDir, "archive", "Data", "TODO")
var persistFile = filepath.Join(persistDir, "queue.json")

func (t *todo) saveToDisk() {
	items := t.workQueue.(*workqueue.DefaultQueue[string]).Snapshot()
	if len(items) == 0 {
		return
	}
	if err := os.MkdirAll(persistDir, 0755); err != nil {
		log.Default().Error("TODO", "mkdir", err)
		return
	}
	data, err := json.Marshal(items)
	if err != nil {
		log.Default().Error("TODO", "marshal", err)
		return
	}
	if err := os.WriteFile(persistFile, data, 0644); err != nil {
		log.Default().Error("TODO", "write", err)
	}
}

func (t *todo) loadFromDisk() {
	data, err := os.ReadFile(persistFile)
	if err != nil {
		return
	}
	var items []string
	if err := json.Unmarshal(data, &items); err != nil {
		return
	}
	for _, item := range items {
		fmt.Printf("TODO item: %s\n", item)
		t.workQueue.Add(item)
	}
}

func New() plugin.Interface {
	return &todo{
		workQueue:  workqueue.New[string](),
		msgChannel: make(chan string, msgChannelBuffer),
		shutdown:   make(chan struct{}),
	}
}
