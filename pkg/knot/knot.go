package knot

import (
	"goldenglow/pkg/node"
)

type Interface interface {
	Trigger() node.Interface
	State() string
}
type knot struct {
	trigger node.Interface
	state   string
}

func (k knot) Trigger() node.Interface {
	return k.trigger
}

func (k knot) State() string {
	return k.state
}

func New(t node.Interface, state string) Interface {
	if t == nil {
		return nil
	}
	return &knot{
		trigger: t,
		state:   state,
	}
}
