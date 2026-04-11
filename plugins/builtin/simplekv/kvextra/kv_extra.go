package kvextra

import (
	"goldenglow/dataGen"
	"goldenglow/executor"
	"goldenglow/executor/checker"
	"goldenglow/lang"
	"goldenglow/plugins"
	"goldenglow/storage"
)

func init() {
	if err := plugins.Subscribe(NewKVExtraPlugin(storage.DefaultRedisRepo())); err != nil {
		panic(err)
	}
}

const (
	pluginName = "kv_extra"

	// Parameter keys
	keyKey   = "key"
	keyValue = "value"
	keyMode  = "mode"

	// Mode values
	modeNotFound      = "not_found"
	modeAlreadyExists = "already_exists"
	modeNotEqual      = "not_equal"
)

type kvExtra struct {
	plugins.Base
	repo storage.RedisRepository
}

func (k *kvExtra) Name() string {
	return pluginName
}

func (k *kvExtra) OnRegisterExecutor(_ executor.Registry) error {
	return nil
}

func (k *kvExtra) OnRegisterChecker(reg checker.Registry) error {
	return reg.Register(pluginName, func(params executor.Parameters) bool {
		if err := executor.Validate(params, keyMode, keyKey); err != nil {
			return false
		}

		mode := params[keyMode]

		switch mode {
		case modeNotFound:
			// Check if key does NOT exist (no value found)
			res, _ := k.repo.HGet(params[keyKey])
			return res == nil

		case modeAlreadyExists:
			// Check if key-value pair already exists
			if err := executor.Validate(params, keyValue); err != nil {
				return false
			}
			res, _ := k.repo.HGet(params[keyKey])
			if res == nil {
				return false
			}
			_, ok := res[params[keyValue]]
			return ok

		case modeNotEqual:
			// Check if value is NOT equal to the parameter value (reversed logic)
			if err := executor.Validate(params, keyValue); err != nil {
				return false
			}
			res, _ := k.repo.HGet(params[keyKey])
			if res == nil {
				return true // If key doesn't exist, it's definitely not equal
			}
			_, ok := res[params[keyValue]]
			return !ok // Return true if NOT found (reversed logic)

		default:
			return false
		}
	})
}

func (k *kvExtra) OnRegisterDataGen(reg dataGen.Registry) error {
	generator := dataGen.NewGenerator(pluginName)

	// Mode 1: Check if key does NOT exist
	generator.Add("not_found", dataGen.SNew(
		"check what is $1",
		"failed to get the value of $1 from Susie 's brain directly",
		dataGen.Parameters{
			keyKey:  "$1",
			keyMode: modeNotFound,
		},
		dataGen.LangTypeChecker,
	))

	// Mode 2: Check if key-value already exists
	generator.Add("already_exists", dataGen.NewLangData(
		[]string{
			"$1 is $2",
		},
		[]string{"the key $1 with the value $2 already exists in Susie's brain"},
		dataGen.Parameters{
			keyKey:   "$1",
			keyValue: "$2",
			keyMode:  modeAlreadyExists,
		},
		dataGen.LangTypeChecker,
	))

	// Mode 3: Check if value is NOT equal
	generator.Add("not_equal", dataGen.SNew(
		"check if $1 is $2",
		"the key $1 with the value $2 does not exist in Susie's brain",
		dataGen.Parameters{
			keyKey:   "$1",
			keyValue: "$2",
			keyMode:  modeNotEqual,
		},
		dataGen.LangTypeChecker,
	))

	return reg.AddGenerator(pluginName, generator)
}

func (k *kvExtra) OnRegisterLang(reg lang.Registry) error {
	return reg.Register(k.Name())
}

func NewKVExtraPlugin(repo storage.RedisRepository) plugins.Item {
	return &kvExtra{
		repo: repo,
	}
}
