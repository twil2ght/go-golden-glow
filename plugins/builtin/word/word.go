package word

import (
	"fmt"
	"goldenglow/dataGen"
	"goldenglow/executor"
	"goldenglow/executor/extractor"
	"goldenglow/lang"
	"goldenglow/plugins"
	"goldenglow/variable"
	"strconv"
	"strings"
)

func init() {
	if err := plugins.Subscribe(NewWordPlugin()); err != nil {
		panic(err)
	}
}

const (
	pluginName = "word"

	// Parameter keys
	keyPhrase = "phrase"
	keyIndex  = "index"
	keyDist   = "dist"
	keyMode   = "mode"

	// Mode values
	modeLength = "length"
	modeWordAt = "word_at"
)

type word struct {
	plugins.Base
}

func (w *word) Name() string {
	return pluginName
}

func (w *word) OnRegisterExecutor(_ executor.Registry) error {
	return nil
}

func (w *word) OnRegisterExtractor(reg extractor.Registry) error {
	return reg.Register(w.Name(), func(params executor.Parameters) (variable.Item, error) {
		if err := executor.Validate(params, keyPhrase, keyMode, keyDist); err != nil {
			return nil, err
		}

		words := strings.Fields(params[keyPhrase])
		mode := params[keyMode]

		switch mode {
		case modeLength:
			// Return the length of the phrase (number of words)
			length := len(words)
			return variable.New(params[keyDist], strconv.Itoa(length)), nil

		case modeWordAt:
			// Return the word at the specified index
			if err := executor.Validate(params, keyIndex); err != nil {
				return nil, err
			}

			index, err := strconv.Atoi(params[keyIndex])
			if err != nil {
				return nil, fmt.Errorf("invalid index '%s': %w", params[keyIndex], err)
			}

			if index < 0 || index >= len(words) {
				return nil, fmt.Errorf("index %d out of range [0, %d)", index, len(words))
			}

			return variable.New(params[keyDist], words[index]), nil

		default:
			return nil, fmt.Errorf("unknown word plugin mode: %s", mode)
		}
	})
}

func (w *word) OnRegisterDataGen(reg dataGen.Registry) error {
	generator := dataGen.NewGenerator(w.Name())
	generator.Add("get_length", dataGen.SNew(
		"check what is the amount of words in phrase $1",
		"",
		dataGen.Parameters{
			keyDist:   "$2",
			keyPhrase: "$1",
			keyMode:   modeLength,
		},
		dataGen.LangTypeExtractor,
	))
	generator.Add("get_length", dataGen.SNew(
		"check what is the word at index $3 in phrase $1",
		"",
		dataGen.Parameters{
			keyDist:   "$2",
			keyPhrase: "$1",
			keyIndex:  "$3",
			keyMode:   modeWordAt,
		},
		dataGen.LangTypeExtractor,
	))
	return reg.AddGenerator(w.Name(), generator)
}

func (w *word) OnRegisterLang(reg lang.Registry) error {
	return reg.Register(w.Name())
}

func NewWordPlugin() plugins.Item {
	return &word{}
}
