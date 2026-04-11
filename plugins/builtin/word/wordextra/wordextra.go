package wordextra

import (
	"goldenglow/dataGen"
	"goldenglow/executor"
	"goldenglow/executor/checker"
	"goldenglow/lang"
	"goldenglow/plugins"
	"strings"
)

func init() {
	if err := plugins.Subscribe(NewWordExtraPlugin()); err != nil {
		panic(err)
	}
}

const (
	pluginName = "word_extra"

	// Parameter keys
	keyPhrase = "phrase"
	keyTarget = "target"
)

type wordExtra struct {
	plugins.Base
}

func (w *wordExtra) Name() string {
	return pluginName
}

func (w *wordExtra) OnRegisterExecutor(_ executor.Registry) error {
	return nil
}

func (w *wordExtra) OnRegisterChecker(reg checker.Registry) error {
	return reg.Register(w.Name(), func(params executor.Parameters) bool {
		if err := executor.Validate(params, keyPhrase, keyTarget); err != nil {
			return false
		}

		phrase := params[keyPhrase]
		target := params[keyTarget]

		// Check if target does NOT exist in phrase (reversed logic)
		return !strings.Contains(phrase, target)
	})
}
func (w *wordExtra) OnRegisterDataGen(reg dataGen.Registry) error {
	generator := dataGen.NewGenerator(w.Name())
	generator.Add("not_contains", dataGen.SNew(
		"check if the word $2 is in phrase $1",
		"the word $2 is not in phrase $1",
		dataGen.Parameters{
			keyPhrase: "$1",
			keyTarget: "$2",
		},
		dataGen.LangTypeChecker,
	))
	return reg.AddGenerator(w.Name(), generator)
}

func (w *wordExtra) OnRegisterLang(reg lang.Registry) error {
	return reg.Register(w.Name())
}

func NewWordExtraPlugin() plugins.Item {
	return &wordExtra{}
}
