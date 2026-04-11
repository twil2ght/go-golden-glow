package kvalreadyexists

import (
	"goldenglow/dataGen"
	"goldenglow/executor"
	"goldenglow/executor/checker"
	"goldenglow/lang"
	"goldenglow/plugins"
	"goldenglow/storage"
)

func init() {
	if err := plugins.Subscribe(NewKVAlreadyExistsPlugin(storage.DefaultRedisRepo())); err != nil {
		panic(err)
	}
}

const (
	pluginName = "kv_already_exists"

	// Parameter keys
	keyKey   = "key"
	keyValue = "value"
)

type kvAlreadyExists struct {
	plugins.Base
	repo storage.RedisRepository
}

func (k *kvAlreadyExists) Name() string {
	return pluginName
}

func (k *kvAlreadyExists) OnRegisterExecutor(_ executor.Registry) error {
	return nil
}

func (k *kvAlreadyExists) OnRegisterChecker(reg checker.Registry) error {
	return reg.Register(pluginName, func(params executor.Parameters) bool {
		if err := executor.Validate(params, keyKey, keyValue); err != nil {
			return false
		}
		res, _ := k.repo.HGet(params[keyKey])
		if res == nil {
			return false
		}
		_, ok := res[params[keyValue]]
		return ok
	})
}

func (k *kvAlreadyExists) OnRegisterDataGen(reg dataGen.Registry) error {
	generator := dataGen.NewGenerator(pluginName)

	generator.Add("set_value", dataGen.NewLangData(
		[]string{
			"$1 is $2",
		},
		[]string{"the key $1 with the value $2 already exists in Susie's brain"},
		dataGen.Parameters{
			keyKey:   "$1",
			keyValue: "$2",
		},
		dataGen.LangTypeChecker,
	))

	return reg.AddGenerator(pluginName, generator)
}

func (k *kvAlreadyExists) OnRegisterLang(reg lang.Registry) error {
	return reg.Register(k.Name())
}

func NewKVAlreadyExistsPlugin(repo storage.RedisRepository) plugins.Item {
	return &kvAlreadyExists{
		repo: repo,
	}
}
