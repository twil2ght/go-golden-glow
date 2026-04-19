package messageQueue

import (
	"context"
	"goldenglow/pkg/registry"
)

type Manager interface {
	Add(name string, provider chan string)
	Start(msgQueue Interface, ctx context.Context)
}
type manager struct {
	items registry.Interface[chan string]
}

func (m *manager) Add(name string, provider chan string) {
	m.items.Register(name, provider)
}

func (m *manager) Start(msgQueue Interface, ctx context.Context) {
	m.items.Range(func(key string, item chan string) bool {
		go func(e chan string, ctx context.Context) {
			for {
				select {
				case <-ctx.Done():
					return
				case msg := <-item:
					msgQueue.Add(msg)
				}
			}
		}(item, ctx)
		return true
	})
	<-ctx.Done()
	msgQueue.Shutdown()
}
func NewManager() Manager {
	return &manager{
		items: registry.New[chan string](),
	}
}
