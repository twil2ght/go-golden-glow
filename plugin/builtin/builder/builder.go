package builder

import (
	"encoding/json"
	"fmt"
	"goldenglow/config"
	"goldenglow/m"
	"goldenglow/pkg/brainsaver"
	"goldenglow/pkg/datagen"
	"goldenglow/pkg/log"
	"goldenglow/pkg/node/handler"
	"goldenglow/plugin"
	"goldenglow/utils"
	"os"
	"path/filepath"
	"strings"
)

func init() {
	plugin.DefaultManager.Register(pluginName, NewBuilderPlugin(brainsaver.DefaultService()))
}

const (
	pluginName = "builder"

	keyValue = "value"
	keyMode  = "mode"
	keyType  = "type"

	KeyName = "name"
	KeyArgs = "args"

	isTesting = false
)

var logger = log.Default()
var (
	// modes
	modeMultiCondition = "multi_condition"
	modeSingleInput    = "single_input"
	modeTemplate       = "template"

	// types
	typeInput  = "input"
	typeOutput = "output"

	// special values
	specialValueClear = "[clear]"
)

type builder struct {
	saver       brainsaver.Service
	mapping     map[string]string
	input       []string
	inputSingle []string
	buildDone   bool
	templates   m.Map[*template]
}

func (b *builder) OnRegisterDataGen(gen datagen.Generator) {
	provider := datagen.NewProvider()
	provider.Add("input", datagen.NewData(
		[]string{fmt.Sprintf("%s says to %s : [input] $1", config.User, config.GG)},
		[]string{},
		map[string]string{
			keyValue: "$1",
			keyMode:  modeMultiCondition,
			keyType:  typeInput,
		},
		datagen.AsExecutor,
	))
	provider.Add("input_v2", datagen.NewData(
		[]string{"[input] $1"},
		[]string{},
		map[string]string{
			keyValue: "$1",
			keyMode:  modeMultiCondition,
			keyType:  typeInput,
		},
		datagen.AsExecutor,
	))
	provider.Add("output", datagen.NewData(
		[]string{fmt.Sprintf("%s says to %s : [output] $1", config.User, config.GG)},
		[]string{},
		map[string]string{
			keyValue: "$1",
			keyMode:  modeMultiCondition,
			keyType:  typeOutput,
		},
		datagen.AsExecutor,
	))
	provider.Add("output_v2", datagen.NewData(
		[]string{"[output] $1"},
		[]string{},
		map[string]string{
			keyValue: "$1",
			keyMode:  modeMultiCondition,
			keyType:  typeOutput,
		},
		datagen.AsExecutor,
	))
	provider.Add("template", datagen.NewData(
		[]string{"[template] $1 @Args $2 @Caller $3"},
		[]string{},
		map[string]string{
			KeyName: "$1",
			KeyArgs: "$2",
			keyMode: modeTemplate,
		},
		datagen.AsExecutor,
	))
	gen.AddProvider(pluginName, provider)
}

func (b *builder) OnRegisterExecutor(reg handler.Executor[handler.ExecuteHandler]) {
	reg.Handlers().Register(pluginName, func(parameters handler.Parameters) {
		var (
			value, _     = parameters.Get(keyValue)
			valueType, _ = parameters.Get(keyType)
			mode, _      = parameters.Get(keyMode)
		)
		if mode == modeTemplate {
			var (
				name, _ = parameters.Get(KeyName)
				args, _ = parameters.Get(KeyArgs)
			)
			b.RunTemplate(name, args)
			return
		}
		_ = b.add(value, valueType, mode)
	})
}

func (b *builder) Init() {
	_ = b.Setup()
}

func (b *builder) Shutdown() {}

func (b *builder) add(value, valueType, mode string) error {
	value = b.mapToPlaceholder(value)
	switch mode {
	case modeSingleInput:
		b.inputSingle = append(b.inputSingle, value)
	}
	if valueType == typeOutput {
		b.buildDone = true
		if value != specialValueClear {
			err := b.build(value)
			if err != nil {
				return err
			}
			return b.buildSingle(value)
		}
	}
	if b.buildDone {
		b.input = nil
		b.inputSingle = nil
		b.buildDone = false
	}
	b.input = append(b.input, value)
	logger.Debug("Builder:input", "input", value)
	return nil
}
func (b *builder) buildSingle(output string) error {
	if len(b.input) == 1 && len(b.inputSingle) > 0 {
		for _, input := range b.inputSingle {
			b.saver.Save(m.ToHash([]string{input}), m.ToHash([]string{output}))
			logger.Debug("Builder:", "input", input, "output", output)
		}
	}
	return nil
}
func (b *builder) build(output string) error {
	if len(b.input) == 0 {
		return fmt.Errorf("no input")
	}
	if !isTesting {
		b.saver.Save(m.ToHash(b.input), m.ToHash([]string{output}))
	}
	logger.Debug("Builder:start build", "inputs", b.input, "output", output)
	return nil
}
func NewBuilderPlugin(saver brainsaver.Service) plugin.Interface {
	return &builder{
		saver:   saver,
		mapping: make(map[string]string),
	}
}

var mappingPath = filepath.Join(utils.RootDir, "config/builder_mapping.json")

func (b *builder) mapToPlaceholder(value string) string {
	var parts = strings.Fields(value)
	for i, part := range parts {
		if placeholder, exists := b.mapping[part]; exists {
			parts[i] = placeholder
		}
	}
	return strings.Join(parts, " ")
}
func (b *builder) ParseTpl(tpl *template, arg m.Map[string]) {
	for _, item := range tpl.Data {
		var inputs, outputs []string
		for _, cmd := range item.Commands {
			cmd = b.mapToPlaceholder(replaceVars(cmd, arg))
			switch {
			case strings.HasPrefix(cmd, "[input] "):
				inputs = append(inputs, strings.TrimPrefix(cmd, "[input] "))
			case strings.HasPrefix(cmd, "[output] "):
				outputs = append(outputs, strings.TrimPrefix(cmd, "[output] "))
			}
		}
		if len(inputs) > 0 && len(outputs) > 0 {
			b.saver.Save(m.ToHash(inputs), m.ToHash(outputs))
			logger.Debug("templateGen: saved container",
				"template", tpl.Name, "inputs", inputs, "outputs", outputs)
		}
	}
}
func (b *builder) RunTemplate(name, stringArgs string) {
	tpl := b.templates[name]
	if tpl == nil {
		logger.Error("template not found", "name", name)
		return
	}
	args, err := parseTemplateArgs(stringArgs)
	if err != nil {
		logger.Error("parse template args error", "name", name, "err", err)
		return
	}
	b.ParseTpl(tpl, args)
}
func (b *builder) Setup() error {
	mappingData, err := os.ReadFile(mappingPath)
	if err != nil {
		return fmt.Errorf("read mapping file: %v", err)
	}
	tpls, err := loadTemplates(templateDir)
	if err != nil {
		return fmt.Errorf("load templates: %v", err)
	}
	b.templates = tpls
	err = json.Unmarshal(mappingData, &b.mapping)
	if err != nil {
		return fmt.Errorf("unmarshal mapping file: %v", err)
	}
	return nil
}
