package word

import (
	"fmt"
	"goldenglow/dataGen"
	"goldenglow/executor"
	"goldenglow/executor/checker"
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
	keyTarget = "target"

	// Mode values
	modeLength  = "length"
	modeWordAt  = "word_at"
	modeIndexOf = "index_of"
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

func (w *word) OnRegisterChecker(reg checker.Registry) error {
	return reg.Register(w.Name(), func(params executor.Parameters) bool {
		if err := executor.Validate(params, keyPhrase, keyTarget); err != nil {
			return false
		}

		phrase := params[keyPhrase]
		target := params[keyTarget]

		// Check if target exists in phrase
		return strings.Contains(phrase, target)
	})
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

			return variable.New(params[keyDist], words[index-1]), nil

		case modeIndexOf:
			// Return the index of the target word/phrase in the phrase
			if err := executor.Validate(params, keyTarget); err != nil {
				return nil, err
			}

			target := params[keyTarget]

			// Search for the target in the words array
			foundIndex := -1
			for i, word := range words {
				if word == target {
					foundIndex = i + 1
					break
				}
			}

			if foundIndex == -1 {
				return nil, fmt.Errorf("target '%s' not found in phrase", target)
			}

			return variable.New(params[keyDist], strconv.Itoa(foundIndex)), nil

		default:
			return nil, fmt.Errorf("unknown word plugin mode: %s", mode)
		}
	})
}

func (w *word) OnRegisterDataGen(reg dataGen.Registry) error {
	generator := dataGen.NewGenerator(w.Name())
	generator.Add("get_length", dataGen.SNew(
		"check what is the amount of words in phrase $1",
		"the amount of words in phrase $1 is $2",
		dataGen.Parameters{
			keyDist:   "$2",
			keyPhrase: "$1",
			keyMode:   modeLength,
		},
		dataGen.LangTypeExtractor,
	))
	generator.Add("get_word_at", dataGen.SNew(
		"check what is the word at index $3 in phrase $1",
		"the word at index $3 in phrase $1 is $2",
		dataGen.Parameters{
			keyDist:   "$2",
			keyPhrase: "$1",
			keyIndex:  "$3",
			keyMode:   modeWordAt,
		},
		dataGen.LangTypeExtractor,
	))
	generator.Add("index_of", dataGen.SNew(
		"check what is the index of the word $3 in phrase $1",
		"the index of the word $3 in phrase $1 is $2",
		dataGen.Parameters{
			keyDist:   "$2",
			keyPhrase: "$1",
			keyTarget: "$3",
			keyMode:   modeIndexOf,
		},
		dataGen.LangTypeExtractor,
	))
	generator.Add("contains", dataGen.SNew(
		"check if the word $2 is in phrase $1",
		"the word $2 is in phrase $1",
		dataGen.Parameters{
			keyPhrase: "$1",
			keyTarget: "$2",
		},
		dataGen.LangTypeChecker,
	))
	return reg.AddGenerator(w.Name(), generator)
}

func (w *word) OnRegisterLang(reg lang.Registry) error {
	return reg.Register(w.Name())
}

func NewWordPlugin() plugins.Item {
	return &word{}
}
