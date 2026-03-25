package builder

import (
	"goldenglow/dataGen"
	"goldenglow/executor"
	"goldenglow/lang"
	"goldenglow/plugins"
)

func init() {
	if err := plugins.Subscribe(NewBuilderPlugin()); err != nil {
		panic(err)
	}
}

const (
	pluginName = "builder"

	keyValue = "value"
	keyMode  = "mode"
	keyType  = "type"
)

var (
	// modes
	modeMultiCondition = "multi_condition"
	modeMultiResult    = "multi_result"

	// types
	typeInput  = "input"
	typeOutput = "output"
)

type builder struct {
	plugins.Base
}

func (b *builder) Name() string {
	return pluginName
}

func (b *builder) add(value, valueType, mode string) {
	//	TODO implement me
}

func (b *builder) OnRegisterExecutor(reg executor.Registry) error {
	return reg.Register(b.Name(), func(params executor.Parameters) error {
		if err := executor.Validate(params, keyValue, keyType, keyMode); err != nil {
			return err
		}

		var (
			value     = params[keyValue]
			mode      = params[keyMode]
			valueType = params[keyType]
		)

		b.add(value, valueType, mode)

		return nil
	})
}

func (b *builder) OnRegisterDataGen(reg dataGen.Registry) error {
	generator := dataGen.NewGenerator(b.Name())
	//TODO implement me:need 4 langItems
	generator.Add(modeMultiCondition, dataGen.New(
		[]string{},
		dataGen.Parameters{},
		dataGen.LangTypeDefault,
	))

	return reg.AddGenerator(b.Name(), generator)
}

func (b *builder) OnRegisterLang(reg lang.Registry) error {
	return reg.Register(pluginName)
}

func NewBuilderPlugin() plugins.Item {
	return &builder{}
}
