package plugins

import (
	"goldenglow/components/preprocessor"
	"goldenglow/components/source"
)

type Base struct{}

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
