package plugins

import (
	"goldenglow/components/preprocessor"
	"goldenglow/components/source"
	"goldenglow/lang"
)

type Base struct {
	name string
}

func (b *Base) OnRegisterLang(langRegistry lang.Registry) error {
	return langRegistry.Register(b.name)
}
func (b *Base) OnRegisterInputSource(sourceRegistry source.Registry) error {
	return nil
}
func (b *Base) OnRegisterPreprocessor(preprocessorRegistry preprocessor.Registry) error {
	return nil
}
func (b *Base) Setup() error {
	return nil
}

func (b *Base) Cleanup() {}

func (b *Base) Name() string {
	return b.name
}
