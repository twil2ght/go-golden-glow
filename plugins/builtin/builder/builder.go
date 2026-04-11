package builder

import (
	"encoding/json"
	"fmt"
	"goldenglow/config"
	"goldenglow/container"
	"goldenglow/dataGen"
	"goldenglow/executor"
	"goldenglow/lang"
	"goldenglow/m"
	"goldenglow/pkg/log"
	"goldenglow/plugins"
	"goldenglow/utils"
	"os"
	"strings"
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
	modeSingleInput    = "single_input"

	// types
	typeInput  = "input"
	typeOutput = "output"
)

type builder struct {
	plugins.Base
	saver       container.Store
	pocketCs2R  []string
	pocketC2Rs  []string
	mapping     map[string]string
	input       []string
	inputSingle []string
	buildDone   bool
}

func (b *builder) Name() string {
	return pluginName
}

func (b *builder) add(value, valueType, mode string) error {
	value = b.mapToPlaceholder(value)
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
func (b *builder) addV2(value, valueType, mode string) error {
	value = b.mapToPlaceholder(value)
	switch mode {
	case modeSingleInput:
		b.inputSingle = append(b.inputSingle, value)
	}
	if valueType == typeOutput {
		b.buildDone = true
		err := b.buildV2(value)
		if err != nil {
			return err
		}
		return b.buildSingle(value)
	}
	if b.buildDone {
		b.input = nil
		b.inputSingle = nil
		b.buildDone = false
	}
	b.input = append(b.input, value)
	logger.Debug("Builder:add input", "input", value)
	return nil
}
func (b *builder) buildSingle(output string) error {
	if len(b.input) == 1 && len(b.inputSingle) > 0 {
		for _, input := range b.inputSingle {
			err := b.saver.Save(m.ToHash([]string{input}), m.ToHash([]string{output}))
			if err != nil {
				return fmt.Errorf("%s save: %v", pluginName, err)
			}
			logger.Debug("Builder:start build single", "input", input, "output", output)
		}
	}
	return nil
}
func (b *builder) buildV2(output string) error {
	if len(b.input) == 0 {
		return fmt.Errorf("no input")
	}
	err := b.saver.Save(m.ToHash(b.input), m.ToHash([]string{output}))
	if err != nil {
		return fmt.Errorf("%s save: %v", pluginName, err)
	}
	logger.Debug("Builder:start build", "inputs", b.input, "output", output)
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

	err := b.saver.Save(m.ToHash(conditions), m.ToHash(result))
	if err != nil {
		return fmt.Errorf("%s save: %v", pluginName, err)
	}
	logger.Debug("Builder:buildMultiCondition", "conditions", conditions, "result", result)
	return nil
}

func (b *builder) buildMultiResult() error {
	defer func() { b.pocketC2Rs = nil }()
	data := b.pocketC2Rs

	if len(data) < 2 {
		return fmt.Errorf("pocketC2Rs length < 2, data: %v", data)
	}

	condition := []string{data[len(data)-1]}
	results := data[:len(data)-1]

	err := b.saver.Save(m.ToHash(condition), m.ToHash(results))
	if err != nil {
		return fmt.Errorf("%s save: %v", pluginName, err)
	}
	logger.Debug("Builder:buildMultiResult", "condition", condition, "results", results)
	return nil
}

func (b *builder) OnRegisterExecutor(reg executor.Registry) error {
	return reg.Register(b.Name(), func(params executor.Parameters) error {
		if err := executor.Validate(params, keyValue, keyType, keyMode); err != nil {
			return err
		}
		//return b.add(params[keyValue], params[keyType], params[keyMode])
		return b.addV2(params[keyValue], params[keyType], params[keyMode])
	})
}

func (b *builder) OnRegisterDataGen(reg dataGen.Registry) error {
	generator := dataGen.NewGenerator(b.Name())

	generator.Add("input_V2", dataGen.SNew(
		fmt.Sprintf("%s says [input] $1 to %s", config.User, config.GG),
		"",
		dataGen.Parameters{
			keyValue: "$1",
			keyMode:  modeMultiCondition,
			keyType:  typeInput,
		},
		dataGen.LangTypeDefault,
	))
	generator.Add("output_V2", dataGen.SNew(
		fmt.Sprintf("%s says [output] $1 to %s", config.User, config.GG),
		"",
		dataGen.Parameters{
			keyValue: "$1",
			keyMode:  modeMultiCondition,
			keyType:  typeOutput,
		},
		dataGen.LangTypeDefault,
	))
	generator.Add("input_single", dataGen.SNew(
		fmt.Sprintf("%s says [input_single] $1 to %s", config.User, config.GG),
		"",
		dataGen.Parameters{
			keyValue: "$1",
			keyMode:  modeMultiCondition,
			keyType:  typeInput,
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
		saver:   saver,
		mapping: make(map[string]string),
	}
}

var mappingPath = utils.RootDir + "/plugins/builtin/builder/mapping.json"

func (b *builder) mapToPlaceholder(value string) string {
	var parts = strings.Fields(value)
	for i, part := range parts {
		if placeholder, exists := b.mapping[part]; exists {
			parts[i] = placeholder
		}
	}
	return strings.Join(parts, " ")
}
func (b *builder) Setup() error {
	mappingData, err := os.ReadFile(mappingPath)
	if err != nil {
		return fmt.Errorf("read mapping file: %v", err)
	}
	err = json.Unmarshal(mappingData, &b.mapping)
	if err != nil {
		return fmt.Errorf("unmarshal mapping file: %v", err)
	}
	logger.Info("Builder:Setup Done", "mapping", b.mapping)
	return nil
}
