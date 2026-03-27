package builder

import (
	"fmt"
	"goldenglow/config"
	"goldenglow/container"
	"goldenglow/dataGen"
	"goldenglow/executor"
	"goldenglow/lang"
	"goldenglow/m"
	"goldenglow/pkg/log"
	"goldenglow/plugins"
)

func init() {
	if err := plugins.Subscribe(NewBuilderPlugin(container.DefaultStore())); err != nil {
		panic(err)
	}
}

const (
	pluginName = "builder"

	keyValue = "value"
	keyMode  = "mode"
	keyType  = "type"
)

var logger = log.Default()
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
	saver      container.Store
	pocketCs2R []string
	pocketC2Rs []string
}

func (b *builder) Name() string {
	return pluginName
}

func (b *builder) add(value, valueType, mode string) error {
	switch mode {
	case modeMultiCondition:
		b.pocketCs2R = append(b.pocketCs2R, value)
	case modeMultiResult:
		b.pocketC2Rs = append(b.pocketC2Rs, value)
	}
	logger.Debug("Builder:add value", "value", value, "valueType", valueType, "mode", mode)
	if valueType == typeOutput {
		return b.build(mode)
	}
	return nil
}

func (b *builder) build(mode string) error {
	switch mode {
	case modeMultiCondition:
		return b.buildMultiCondition()
	case modeMultiResult:
		return b.buildMultiResult()
	default:
		return fmt.Errorf("unsupported mode: %s", mode)
	}
}

func (b *builder) buildMultiCondition() error {
	defer func() { b.pocketCs2R = nil }()
	data := b.pocketCs2R

	if len(data) < 2 {
		return fmt.Errorf("pocketCs2R length < 2, data: %v", data)
	}

	conditions := data[:len(data)-1]
	result := []string{data[len(data)-1]}

	return b.saver.Save(m.ToHash(conditions), m.ToHash(result))
}

func (b *builder) buildMultiResult() error {
	defer func() { b.pocketC2Rs = nil }()
	data := b.pocketC2Rs

	if len(data) < 2 {
		return fmt.Errorf("pocketC2Rs length < 2, data: %v", data)
	}

	condition := []string{data[len(data)-1]}
	results := data[:len(data)-1]

	return b.saver.Save(m.ToHash(condition), m.ToHash(results))
}

func (b *builder) OnRegisterExecutor(reg executor.Registry) error {
	return reg.Register(b.Name(), func(params executor.Parameters) error {
		if err := executor.Validate(params, keyValue, keyType, keyMode); err != nil {
			return err
		}
		return b.add(params[keyValue], params[keyType], params[keyMode])
	})
}

func (b *builder) OnRegisterDataGen(reg dataGen.Registry) error {
	generator := dataGen.NewGenerator(b.Name())
	generator.Add("input_Cs", dataGen.New(
		[]string{fmt.Sprintf("%s says [input_Cs] $1 to %s", config.User, config.GG)},
		dataGen.Parameters{
			keyValue: "$1",
			keyMode:  modeMultiCondition,
			keyType:  typeInput,
		},
		dataGen.LangTypeDefault,
	))
	generator.Add("output_Cs", dataGen.New(
		[]string{fmt.Sprintf("%s says [output_Cs] $1 to %s", config.User, config.GG)},
		dataGen.Parameters{
			keyValue: "$1",
			keyMode:  modeMultiCondition,
			keyType:  typeOutput,
		},
		dataGen.LangTypeDefault,
	))
	generator.Add("input_Rs", dataGen.New(
		[]string{fmt.Sprintf("%s says [input_Rs] $1 to %s", config.User, config.GG)},
		dataGen.Parameters{
			keyValue: "$1",
			keyMode:  modeMultiResult,
			keyType:  typeInput,
		},
		dataGen.LangTypeDefault,
	))
	generator.Add("output_Rs", dataGen.New(
		[]string{fmt.Sprintf("%s says [output_Rs] $1 to %s", config.User, config.GG)},
		dataGen.Parameters{
			keyValue: "$1",
			keyMode:  modeMultiResult,
			keyType:  typeOutput,
		},
		dataGen.LangTypeDefault,
	))

	return reg.AddGenerator(b.Name(), generator)
}

func (b *builder) OnRegisterLang(reg lang.Registry) error {
	return reg.Register(pluginName)
}

func NewBuilderPlugin(saver container.Store) plugins.Item {
	return &builder{
		saver: saver,
	}
}
