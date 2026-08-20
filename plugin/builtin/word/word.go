package word

import (
	"goldenglow/m"
	"goldenglow/pkg/datagen"
	"goldenglow/pkg/node/handler"
	"goldenglow/pkg/variable"
	"goldenglow/plugin"
	"strconv"
	"strings"
)

func init() {
	plugin.DefaultManager.Register(name, NewWordPlugin())
}

const (
	name = "word"

	// Parameter keys — phrase-level
	keyPhrase = "phrase"
	keyIndex  = "index"
	keyDist   = "dist"
	keyMode   = "mode"
	keyTarget = "target"
	keyCount  = "count"

	// Parameter keys — character-level
	keyWord   = "word"
	keyPrefix = "prefix"
	keySuffix = "suffix"

	// Mode values — phrase-level extractor
	modePhraseLength     = "length"
	modeWordAt           = "word_at"
	modePhraseIndexOf    = "index_of"
	modeGetLast_N_Words  = "truncate_head"
	modeGetFirst_N_Words = "truncate_tail"

	// Mode values — character-level extractor
	modeCharLength  = "char_length"
	modeCharAt      = "char_at"
	modeCharIndexOf = "char_index_of"
	modeAddPrefix   = "add_prefix"
	modeAddSuffix   = "add_suffix"

	// Mode values — checker
	modeStartsWithPhrase = "starts_with_phrase"
	modeStartsWith       = "starts_with"
	modeEndsWith         = "ends_with"
)

type word struct{}

func (w *word) OnRegisterChecker(reg handler.Executor[handler.CheckHandler]) {
	reg.Handlers().Register(name, func(parameters handler.Parameters) bool {
		mode, _ := parameters.Get(keyMode)
		target, err := parameters.Get(keyTarget)
		if err != nil {
			return false
		}

		// Character-level checks (use keyWord)
		if word, err := parameters.Get(keyWord); err == nil {
			switch mode {
			case modeStartsWith:
				return strings.HasPrefix(word, target)
			case modeEndsWith:
				return strings.HasSuffix(word, target)
			}
			return false
		}

		// Phrase-level checks (use keyPhrase)
		if phrase, err := parameters.Get(keyPhrase); err == nil {
			switch mode {
			case modeStartsWithPhrase:
				return strings.HasPrefix(phrase, target)
			default:
				return strings.Contains(phrase, target)
			}
		}

		return false
	})
}

func (w *word) OnRegisterExtractor(reg handler.Executor[handler.ExtractorHandler]) {
	reg.Handlers().Register(name, func(parameters handler.Parameters) variable.ValueMap {
		mode, err := parameters.Get(keyMode)
		if err != nil {
			return nil
		}

		// Character-level extractors (use keyWord)
		if word, err := parameters.Get(keyWord); err == nil {
			switch mode {
			case modeCharLength:
				length := len(word)
				return variable.NewValueMap(m.Hash{strconv.Itoa(length): struct{}{}})

			case modeCharAt:
				indexStr, err := parameters.Get(keyIndex)
				if err != nil {
					return nil
				}
				index, err := strconv.Atoi(indexStr)
				if err != nil {
					return nil
				}
				if index < 0 || index >= len(word) {
					return nil
				}
				return variable.NewValueMap(m.Hash{string(word[index]): struct{}{}})

			case modeCharIndexOf:
				target, err := parameters.Get(keyTarget)
				if err != nil {
					return nil
				}
				foundIndex := strings.Index(word, target)
				if foundIndex == -1 {
					return nil
				}
				return variable.NewValueMap(m.Hash{strconv.Itoa(foundIndex): struct{}{}})

			case modeAddPrefix:
				prefix, err := parameters.Get(keyPrefix)
				if err != nil {
					return nil
				}
				return variable.NewValueMap(m.Hash{prefix + word: struct{}{}})

			case modeAddSuffix:
				suffix, err := parameters.Get(keySuffix)
				if err != nil {
					return nil
				}
				return variable.NewValueMap(m.Hash{word + suffix: struct{}{}})
			}

			return nil
		}

		// Phrase-level extractors (use keyPhrase)
		phrase, err := parameters.Get(keyPhrase)
		if err != nil {
			return nil
		}
		words := strings.Fields(phrase)

		switch mode {
		case modePhraseLength:
			length := len(words)
			return variable.NewValueMap(m.Hash{strconv.Itoa(length): struct{}{}})

		case modeWordAt:
			indexStr, err := parameters.Get(keyIndex)
			if err != nil {
				return nil
			}
			index, err := strconv.Atoi(indexStr)
			if err != nil {
				return nil
			}
			if index < 1 || index > len(words) {
				return nil
			}
			return variable.NewValueMap(m.Hash{words[index-1]: struct{}{}})

		case modePhraseIndexOf:
			target, err := parameters.Get(keyTarget)
			if err != nil {
				return nil
			}
			foundIndex := -1
			for i, w := range words {
				if w == target {
					foundIndex = i + 1
					break
				}
			}
			if foundIndex == -1 {
				return nil
			}
			return variable.NewValueMap(m.Hash{strconv.Itoa(foundIndex): struct{}{}})

		case modeGetLast_N_Words:
			countStr, err := parameters.Get(keyCount)
			if err != nil {
				return nil
			}
			count, err := strconv.Atoi(countStr)
			if err != nil {
				return nil
			}
			if count < 0 || count >= len(words) {
				return nil
			}
			truncated := words[count:]
			return variable.NewValueMap(m.Hash{strings.Join(truncated, " "): struct{}{}})

		case modeGetFirst_N_Words:
			countStr, err := parameters.Get(keyCount)
			if err != nil {
				return nil
			}
			count, err := strconv.Atoi(countStr)
			if err != nil {
				return nil
			}
			if count < 0 || count >= len(words) {
				return nil
			}
			truncated := words[:len(words)-count]
			return variable.NewValueMap(m.Hash{strings.Join(truncated, " "): struct{}{}})

		default:
			return nil
		}
	})
}

func (w *word) OnRegisterDataGen(gen datagen.Generator) {
	provider := datagen.NewProvider()

	// Phrase-level extractors
	provider.Add("get_length", datagen.NewData(
		[]string{"[phrase] [len] $1 @Caller $3"},
		[]string{"[phrase] [len] $3 @ $1 -> $2"},
		map[string]string{
			keyDist:   "$2",
			keyPhrase: "$1",
			keyMode:   modePhraseLength,
		},
		datagen.AsExtractor,
	))
	provider.Add("get_word_at", datagen.NewData(
		[]string{"[phrase] get $1 @ $3"},
		[]string{"[phrase] $1 @ $3 -> $2"},
		map[string]string{
			keyDist:   "$2",
			keyPhrase: "$1",
			keyIndex:  "$3",
			keyMode:   modeWordAt,
		},
		datagen.AsExtractor,
	))
	provider.Add("index_of", datagen.NewData(
		[]string{"[phrase] IndexOf $1 # $3"},
		[]string{"[phrase] IndexOf $1 # $3 -> $2"},
		map[string]string{
			keyDist:   "$2",
			keyPhrase: "$1",
			keyTarget: "$3",
			keyMode:   modePhraseIndexOf,
		},
		datagen.AsExtractor,
	))
	provider.Add("truncate_head", datagen.NewData(
		[]string{"[phrase] LastN $1 # $3"},
		[]string{"[phrase] LastN $1 # $3 -> $2"},
		map[string]string{
			keyDist:   "$2",
			keyPhrase: "$1",
			keyCount:  "$3",
			keyMode:   modeGetLast_N_Words,
		},
		datagen.AsExtractor,
	))
	provider.Add("truncate_tail", datagen.NewData(
		[]string{"[phrase] FirstN $1 # $3"},
		[]string{"[phrase] FirstN $1 # $3 -> $2"},
		map[string]string{
			keyDist:   "$2",
			keyPhrase: "$1",
			keyCount:  "$3",
			keyMode:   modeGetFirst_N_Words,
		},
		datagen.AsExtractor,
	))

	provider.Add("starts_with_phrase", datagen.NewData(
		[]string{"[phrase] check HasPrefix $1 # $2"},
		[]string{"[phrase] HasPrefix $1 # $2"},
		map[string]string{
			keyPhrase: "$1",
			keyTarget: "$2",
			keyMode:   modeStartsWithPhrase,
		},
		datagen.AsChecker,
	))

	provider.Add("get_char_length", datagen.NewData(
		[]string{"[word] len $1"},
		[]string{"[word] len $1 -> $2"},
		map[string]string{
			keyDist: "$2",
			keyWord: "$1",
			keyMode: modeCharLength,
		},
		datagen.AsExtractor,
	))
	provider.Add("get_char_at", datagen.NewData(
		[]string{"[word] get $1 @ $3"},
		[]string{"[word] $1 @ $3 -> $2"},
		map[string]string{
			keyDist:  "$2",
			keyWord:  "$1",
			keyIndex: "$3",
			keyMode:  modeCharAt,
		},
		datagen.AsExtractor,
	))
	provider.Add("char_index_of", datagen.NewData(
		[]string{"[word] IndexOf $1 # $3"},
		[]string{"[word] IndexOf $3 # $1 -> $2"},
		map[string]string{
			keyDist:   "$2",
			keyWord:   "$1",
			keyTarget: "$3",
			keyMode:   modeCharIndexOf,
		},
		datagen.AsExtractor,
	))

	provider.Add("starts_with", datagen.NewData(
		[]string{"[word] check HasPrefix $1 # $2"},
		[]string{"[word] HasPrefix $1 # $2"},
		map[string]string{
			keyWord:   "$1",
			keyTarget: "$2",
			keyMode:   modeStartsWith,
		},
		datagen.AsChecker,
	))
	provider.Add("ends_with", datagen.NewData(
		[]string{"[word] check HasSuffix $1 # $2"},
		[]string{"[word] HasSuffix $1 # $2"},
		map[string]string{
			keyWord:   "$1",
			keyTarget: "$2",
			keyMode:   modeEndsWith,
		},
		datagen.AsChecker,
	))

	provider.Add("add_prefix", datagen.NewData(
		[]string{"[word] [AddPrefix] $1 # $2 @Caller $4"},
		[]string{"[word] [AddPrefix] $4 @ $1 # $2 -> $3"},
		map[string]string{
			keyDist:   "$3",
			keyWord:   "$1",
			keyPrefix: "$2",
			keyMode:   modeAddPrefix,
		},
		datagen.AsExtractor,
	))
	provider.Add("add_suffix", datagen.NewData(
		[]string{"[word] [AddSuffix] $1 # $2 @Caller $4"},
		[]string{"[word] [AddSuffix] $4 @ $1 # $2 -> $3"},
		map[string]string{
			keyDist:   "$3",
			keyWord:   "$1",
			keySuffix: "$2",
			keyMode:   modeAddSuffix,
		},
		datagen.AsExtractor,
	))

	gen.AddProvider(name, provider)
}

func (w *word) Init() {}

func (w *word) Shutdown() {}

func NewWordPlugin() plugin.Interface {
	return &word{}
}
