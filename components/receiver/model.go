package receiver

import (
	"goldenglow/components"
	"goldenglow/m"
)

type Receiver func() chan<- components.Message

type Registry interface {
	Register(tag string, subscribeTo m.Hash, r Receiver) error
	Subscriptions() m.Hash
	Start()
}
type RegisterItem interface {
	OnRegisterReceiver(reg Registry) error
}
