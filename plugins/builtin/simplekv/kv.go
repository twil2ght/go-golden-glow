package simplekv

import (
	"fmt"
	"goldenglow/dataGen"
	"goldenglow/executor"
	"goldenglow/executor/checker"
	"goldenglow/executor/extractor"
	"goldenglow/lang"
	"goldenglow/plugins"
	"goldenglow/storage"
	"goldenglow/variable"
)

func init() {
	if err := plugins.Subscribe(NewSimpleKV(storage.DefaultRedisRepo())); err != nil {
		panic(err)
	}
}

const (
	PluginName = "simple_kv"

	// parameter keys
	keyKey        = "key"
	keyValue      = "value"
	keyExpiration = "expiration"
	keyDist       = "dist"
)

type simpleKV struct {
	plugins.Base
	repo storage.RedisRepository
}

func (s *simpleKV) OnRegisterExtractor(reg extractor.Registry) error {
	return reg.Register(PluginName, func(params executor.Parameters) (variable.Item, error) {
		if err := executor.Validate(params, keyKey, keyDist); err != nil {
			return nil, err
		}
		dist := params[keyDist]
		hash, err := s.repo.HGet(params[keyKey])
		if err != nil {
			return nil, fmt.Errorf("failed to get hash: %w", err)
		}
		var value string
		for v := range hash {
			value = v
			break
		}
		return variable.NewValueMap(dist, value, hash), nil
	})
}

func (s *simpleKV) OnRegisterChecker(reg checker.Registry) error {
	return reg.Register(PluginName, func(params executor.Parameters) bool {
		if err := executor.Validate(params, keyKey, keyValue); err != nil {
			return false
		}

		res, _ := s.repo.HGet(params[keyKey])
		if res == nil {
			return false
		}
		if _, ok := res[params[keyValue]]; !ok {
			return false
		}
		return true
	})
}

func (s *simpleKV) Name() string {
	return PluginName
}

func (s *simpleKV) OnRegisterExecutor(reg executor.Registry) error {
	return reg.Register(PluginName, func(params executor.Parameters) error {
		if err := executor.Validate(params, keyKey, keyValue); err != nil {
			return err
		}
		return s.repo.Set(params[keyKey], params[keyValue], params[keyExpiration])
	})
}

func (s *simpleKV) OnRegisterDataGen(reg dataGen.Registry) error {
	generator := dataGen.NewGenerator(PluginName)

	// Get operation - extractor pattern
	generator.Add("get_value", dataGen.SNew(
		"check what is $1",
		"$1 is $2",
		dataGen.Parameters{
			keyKey:  "$1",
			keyDist: "$2",
		},
		dataGen.LangTypeExtractor,
	))

	//generator.Add("set_value", dataGen.NewLangData(
	//	[]string{
	//		"$1 is $2",
	//	},
	//	[]string{""},
	//	dataGen.Parameters{
	//		keyKey:   "$1",
	//		keyValue: "$2",
	//	},
	//	dataGen.LangTypeDefault,
	//))
	generator.Add("set_value", dataGen.NewLangData(
		[]string{
			"$1 is $2",
			"the key $1 with the value $2 does not exist in Susie's brain directly",
			"the phrase $1 is $2 refers to making attributions",
		},
		[]string{""},
		dataGen.Parameters{
			keyKey:   "$1",
			keyValue: "$2",
		},
		dataGen.LangTypeDefault,
	))

	// Set operation - executor pattern
	//generator.Add("set_value_with_expiration", dataGen.NewLangData(
	//	[]string{
	//		"$1 is $2",
	//		"the timeliness of $1 is $3",
	//	},
	//	[]string{""},
	//	dataGen.Parameters{
	//		keyKey:        "$1",
	//		keyValue:      "$2",
	//		keyExpiration: "$3",
	//	},
	//	dataGen.LangTypeDefault,
	//))
	generator.Add("set_value_with_expiration", dataGen.NewLangData(
		[]string{
			"$1 is $2",
			"the key $1 with the value $2 does not exist in Susie's brain directly",
			"the phrase $1 is $2 refers to making attributions",
			"the timeliness of $1 is $3",
		},
		[]string{""},
		dataGen.Parameters{
			keyKey:        "$1",
			keyValue:      "$2",
			keyExpiration: "$3",
		},
		dataGen.LangTypeDefault,
	))
	// Check if key exists - checker pattern
	generator.Add("check_key_exists", dataGen.SNew(
		"check if $1 is $2",
		"$1 is $2",
		dataGen.Parameters{
			keyKey:   "$1",
			keyValue: "$2",
		},
		dataGen.LangTypeChecker,
	))

	return reg.AddGenerator(PluginName, generator)
}

func (s *simpleKV) OnRegisterLang(reg lang.Registry) error {
	return reg.Register(s.Name())
}

func NewSimpleKV(repo storage.RedisRepository) plugins.Item {
	return &simpleKV{
		repo: repo,
	}
}
