package plugins

import (
	"goldenglow/lang"
)

type base struct {
	name string
}

func (b *base) OnRegisterLang(langRegistry lang.Registry) error {
	return langRegistry.Register(b.name)
}

func (b *base) Setup() error {
	return nil
}

func (b *base) Cleanup() {}

func (b *base) Name() string {
	return b.name
}
