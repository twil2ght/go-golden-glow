package pinkcat

import (
	"fmt"
	"goldenglow/node"
	"goldenglow/pkg/validator"
	"sync"
)

//TODO 大重构

type Decorator func(string) string
type Runner interface {
	Run(node node.Item) error
}
type Opt func(*Engine)
type Engine struct {
	userID      int
	runner      Runner
	IsRunning   bool
	RunningMu   *sync.Mutex
	Decorator   Decorator
	Validator   Decorator
	nodeFactory node.Factory
}

func Single(userID int, options ...Opt) *Engine {
	engine := &Engine{
		userID:    userID,
		RunningMu: &sync.Mutex{},
		Validator: validator.ValidateUserInput,
	}
	engine.SetDecorator(func(s string) string {
		return fmt.Sprintf("[say] %d to [GG] : %s //", engine.userID, s)
	})
	for _, opt := range options {
		if opt != nil {
			opt(engine)
		}
	}
	return engine
}
func (e *Engine) RunningState() bool {
	return e.IsRunning
}
func (e *Engine) SetRunningState(s bool) {
	e.IsRunning = s
}
func (e *Engine) GetLock() *sync.Mutex {
	return e.RunningMu
}
func (e *Engine) GetDecorator() Decorator {
	return e.Decorator
}
func (e *Engine) GetValidator() Decorator {
	return e.Validator
}
func (e *Engine) GetUserID() int {
	return e.userID
}
func (e *Engine) Run(input string) error {
	input = e.ValidateInput(input)
	input = e.DecorateInput(input)
	initialNode, err := e.nodeFactory.New(input)
	if err != nil {
		return err
	}
	return e.runner.Run(initialNode)
}
func (e *Engine) SetDecorator(d Decorator) {
	if d == nil {
		return
	}
	e.Decorator = d
}
func (e *Engine) SetValidator(v Decorator) {
	if v == nil {
		return
	}
	e.Validator = v
}

// DecorateInput generate input with Headset: [say] user to [GG] : input
func (e *Engine) DecorateInput(input string) string {
	if e.Decorator != nil {
		return e.Decorator(input)
	}
	return input
}
func (e *Engine) ValidateInput(input string) string {
	if e.Validator != nil {
		return e.Validator(input)
	}
	return input
}
