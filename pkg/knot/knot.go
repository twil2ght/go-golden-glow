package knot

import (
	"errors"
	"goldenglow/node"
)

type Item interface {
	Trigger() node.Item
	State() string
	SetState(string)
}
type knot struct {
	trigger node.Item
	state   string
}

func (k knot) Trigger() node.Item {
	return k.trigger
}

func (k knot) State() string {
	return k.state
}

func (k knot) SetState(s string) {
	k.state = s
}

func New(t node.Item) (Item, error) {
	if t == nil {
		return nil, errors.New("NewKnot: trigger==nil")
	}
	return &knot{trigger: t}, nil
}
