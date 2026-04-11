package kvnotfound

import (
	"goldenglow/dataGen"
	"goldenglow/executor"
	"goldenglow/executor/checker"
	"goldenglow/lang"
	"goldenglow/plugins"
	"goldenglow/storage"
)

func init() {
	if err := plugins.Subscribe(NewKVNotFoundPlugin(storage.DefaultRedisRepo())); err != nil {
		panic(err)
	}
}

const (
	pluginName = "kv_not_found"

	// Parameter keys
	keyKey = "key"
)

type kvNotFound struct {
	plugins.Base
	repo storage.RedisRepository
}

func (k *kvNotFound) Name() string {
	return pluginName
}

func (k *kvNotFound) OnRegisterExecutor(_ executor.Registry) error {
	return nil
}

func (k *kvNotFound) OnRegisterChecker(reg checker.Registry) error {
	return reg.Register(pluginName, func(params executor.Parameters) bool {
		if err := executor.Validate(params, keyKey); err != nil {
			return false
		}
		res, _ := k.repo.HGet(params[keyKey])
		return res == nil
	})
}
func (k *kvNotFound) OnRegisterDataGen(reg dataGen.Registry) error {
	generator := dataGen.NewGenerator(pluginName)

	generator.Add("get_value", dataGen.SNew(
		"check what is $1",
		"failed to get the value of $1 from Susie 's brain directly",
		dataGen.Parameters{
			keyKey: "$1",
		},
		dataGen.LangTypeChecker,
	))
	return reg.AddGenerator(pluginName, generator)
}

func (k *kvNotFound) OnRegisterLang(reg lang.Registry) error {
	return reg.Register(k.Name())
}

func NewKVNotFoundPlugin(repo storage.RedisRepository) plugins.Item {
	return &kvNotFound{
		repo: repo,
	}
}
