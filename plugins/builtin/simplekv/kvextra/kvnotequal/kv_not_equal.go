package kvnotequal

import (
	"goldenglow/dataGen"
	"goldenglow/executor"
	"goldenglow/executor/checker"
	"goldenglow/lang"
	"goldenglow/plugins"
	"goldenglow/storage"
)

func init() {
	if err := plugins.Subscribe(NewKVNotEqualPlugin(storage.DefaultRedisRepo())); err != nil {
		panic(err)
	}
}

const (
	pluginName = "kv_not_equal"

	// Parameter keys
	keyKey   = "key"
	keyValue = "value"
)

type kvNotEqual struct {
	plugins.Base
	repo storage.RedisRepository
}

func (k *kvNotEqual) Name() string {
	return pluginName
}

func (k *kvNotEqual) OnRegisterExecutor(_ executor.Registry) error {
	return nil
}

func (k *kvNotEqual) OnRegisterChecker(reg checker.Registry) error {
	return reg.Register(pluginName, func(params executor.Parameters) bool {
		if err := executor.Validate(params, keyKey, keyValue); err != nil {
			return false
		}
		res, _ := k.repo.HGet(params[keyKey])
		if res == nil {
			return true // If key doesn't exist, it's definitely not equal
		}
		_, ok := res[params[keyValue]]
		return !ok // Return true if NOT found (reversed logic)
	})
}

func (k *kvNotEqual) OnRegisterDataGen(reg dataGen.Registry) error {
	generator := dataGen.NewGenerator(pluginName)

	generator.Add("check_key_exists", dataGen.SNew(
		"check if $1 is $2",
		"the key $1 with the value $2 does not exist in Susie's brain",
		dataGen.Parameters{
			keyKey:   "$1",
			keyValue: "$2",
		},
		dataGen.LangTypeChecker,
	))

	return reg.AddGenerator(pluginName, generator)
}

func (k *kvNotEqual) OnRegisterLang(reg lang.Registry) error {
	return reg.Register(k.Name())
}

func NewKVNotEqualPlugin(repo storage.RedisRepository) plugins.Item {
	return &kvNotEqual{
		repo: repo,
	}
}
