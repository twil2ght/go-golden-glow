package wordwise

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
	if err := plugins.Subscribe(NewWordWisePlugin()); err != nil {
		panic(err)
	}
}

const (
	pluginName = "word_wise"

	// Parameter keys
	keyWord   = "word"
	keyIndex  = "index"
	keyDist   = "dist"
	keyMode   = "mode"
	keyTarget = "target"
	keyPrefix = "prefix"
	keySuffix = "suffix"

	// Mode values
	modeLength     = "length"
	modeCharAt     = "char_at"
	modeIndexOf    = "index_of"
	modeStartsWith = "starts_with"
	modeEndsWith   = "ends_with"
	modeAddPrefix  = "add_prefix"
	modeAddSuffix  = "add_suffix"
)

type wordWise struct {
	plugins.Base
}

func (w *wordWise) Name() string {
	return pluginName
}

func (w *wordWise) OnRegisterExecutor(_ executor.Registry) error {
	return nil
}

func (w *wordWise) OnRegisterChecker(reg checker.Registry) error {
	return reg.Register(w.Name(), func(params executor.Parameters) bool {
		if err := executor.Validate(params, keyWord, keyMode); err != nil {
			return false
		}

		word := params[keyWord]
		mode := params[keyMode]

		switch mode {
		case modeStartsWith:
			if err := executor.Validate(params, keyTarget); err != nil {
				return false
			}
			return strings.HasPrefix(word, params[keyTarget])

		case modeEndsWith:
			if err := executor.Validate(params, keyTarget); err != nil {
				return false
			}
			return strings.HasSuffix(word, params[keyTarget])

		default:
			return false
		}
	})
}

func (w *wordWise) OnRegisterExtractor(reg extractor.Registry) error {
	return reg.Register(w.Name(), func(params executor.Parameters) (variable.Item, error) {
		if err := executor.Validate(params, keyWord, keyMode, keyDist); err != nil {
			return nil, err
		}

		word := params[keyWord]
		mode := params[keyMode]

		switch mode {
		case modeLength:
			// Return the length of the word (number of characters)
			length := len(word)
			return variable.New(params[keyDist], strconv.Itoa(length)), nil

		case modeCharAt:
			// Return the character at the specified index
			if err := executor.Validate(params, keyIndex); err != nil {
				return nil, err
			}

			index, err := strconv.Atoi(params[keyIndex])
			if err != nil {
				return nil, fmt.Errorf("invalid index '%s': %w", params[keyIndex], err)
			}

			if index < 0 || index >= len(word) {
				return nil, fmt.Errorf("index %d out of range [0, %d)", index, len(word))
			}

			return variable.New(params[keyDist], string(word[index])), nil

		case modeIndexOf:
			// Return the index of the target character/substring in the word
			if err := executor.Validate(params, keyTarget); err != nil {
				return nil, err
			}

			target := params[keyTarget]
			foundIndex := strings.Index(word, target)

			if foundIndex == -1 {
				return nil, fmt.Errorf("target '%s' not found in word", target)
			}

			return variable.New(params[keyDist], strconv.Itoa(foundIndex)), nil

		case modeAddPrefix:
			// Generate a new word by adding a prefix
			if err := executor.Validate(params, keyPrefix); err != nil {
				return nil, err
			}

			prefix := params[keyPrefix]
			newWord := prefix + word
			return variable.New(params[keyDist], newWord), nil

		case modeAddSuffix:
			// Generate a new word by adding a suffix
			if err := executor.Validate(params, keySuffix); err != nil {
				return nil, err
			}

			suffix := params[keySuffix]
			newWord := word + suffix
			return variable.New(params[keyDist], newWord), nil

		default:
			return nil, fmt.Errorf("unknown word_wise plugin mode: %s", mode)
		}
	})
}

func (w *wordWise) OnRegisterDataGen(reg dataGen.Registry) error {
	generator := dataGen.NewGenerator(w.Name())

	// Extractor: Get word length (character count)
	generator.Add("get_length", dataGen.SNew(
		"check what is the amount of characters in word $1",
		"the amount of characters in word $1 is $2",
		dataGen.Parameters{
			keyDist: "$2",
			keyWord: "$1",
			keyMode: modeLength,
		},
		dataGen.LangTypeExtractor,
	))

	// Extractor: Get character at index
	generator.Add("get_char_at", dataGen.SNew(
		"check what is the character at index $3 in word $1",
		"the character at index $3 in word $1 is $2",
		dataGen.Parameters{
			keyDist:  "$2",
			keyWord:  "$1",
			keyIndex: "$3",
			keyMode:  modeCharAt,
		},
		dataGen.LangTypeExtractor,
	))

	// Extractor: Get index of character/substring
	generator.Add("index_of", dataGen.SNew(
		"check what is the index of the character $3 in word $1",
		"the index of the character $3 in word $1 is $2",
		dataGen.Parameters{
			keyDist:   "$2",
			keyWord:   "$1",
			keyTarget: "$3",
			keyMode:   modeIndexOf,
		},
		dataGen.LangTypeExtractor,
	))

	// Checker: Check if word starts with target
	generator.Add("starts_with", dataGen.SNew(
		"check if the word $1 starts with $2",
		"the word $1 starts with $2",
		dataGen.Parameters{
			keyWord:   "$1",
			keyTarget: "$2",
			keyMode:   modeStartsWith,
		},
		dataGen.LangTypeChecker,
	))

	// Checker: Check if word ends with target
	generator.Add("ends_with", dataGen.SNew(
		"check if the word $1 ends with $2",
		"the word $1 ends with $2",
		dataGen.Parameters{
			keyWord:   "$1",
			keyTarget: "$2",
			keyMode:   modeEndsWith,
		},
		dataGen.LangTypeChecker,
	))

	// Extractor: Generate new word by adding prefix
	generator.Add("add_prefix", dataGen.SNew(
		"check what is the word formed by adding prefix $2 to word $1",
		"the word formed by adding prefix $2 to word $1 is $3",
		dataGen.Parameters{
			keyDist:   "$3",
			keyWord:   "$1",
			keyPrefix: "$2",
			keyMode:   modeAddPrefix,
		},
		dataGen.LangTypeExtractor,
	))

	// Extractor: Generate new word by adding suffix
	generator.Add("add_suffix", dataGen.SNew(
		"check what is the word formed by adding suffix $2 to word $1",
		"the word formed by adding suffix $2 to word $1 is $3",
		dataGen.Parameters{
			keyDist:   "$3",
			keyWord:   "$1",
			keySuffix: "$2",
			keyMode:   modeAddSuffix,
		},
		dataGen.LangTypeExtractor,
	))

	return reg.AddGenerator(w.Name(), generator)
}

func (w *wordWise) OnRegisterLang(reg lang.Registry) error {
	return reg.Register(w.Name())
}

func NewWordWisePlugin() plugins.Item {
	return &wordWise{}
}
