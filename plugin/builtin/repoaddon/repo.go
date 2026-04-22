package repoaddon

import (
	"goldenglow/pkg/datagen"
	"goldenglow/pkg/log"
	"goldenglow/pkg/node/handler"
	"goldenglow/plugin"
	"goldenglow/storage"
	"goldenglow/variable"
)

func init() {
	plugin.DefaultManager.Register(name, New(storage.DefaultRedisRepo()))
}

const (
	name     = "repo_addon"
	nameFail = "repo_addon_fail"
	testing  = true
	// parameter keys
	keyKey        = "key"
	keyValue      = "value"
	keyExpiration = "expiration"
	keyDist       = "dist"
)

type addon struct {
	repo storage.RedisRepository
}

func (s *addon) Init() {}

func (s *addon) Shutdown() {}

func (s *addon) OnRegisterDataGen(gen datagen.Generator) {
	provider := datagen.NewProvider()
	providerFail := datagen.NewProvider()
	provider.Add("get_value", datagen.NewData(
		[]string{"[repo] get $1"},
		[]string{"[repo] $1 -> $2"},
		map[string]string{
			keyKey:  "$1",
			keyDist: "$2",
		},
		datagen.AsExtractor,
	))
	provider.Add("set_value", datagen.NewData(
		[]string{"[repo] set $1 -> $2"},
		[]string{},
		map[string]string{
			keyKey:   "$1",
			keyValue: "$2",
		},
		datagen.AsExecutor,
	))
	provider.Add("set_value_with_expiration", datagen.NewData(
		[]string{"[repo] set $1 -> $2 @ $3"},
		[]string{},
		map[string]string{
			keyKey:        "$1",
			keyValue:      "$2",
			keyExpiration: "$3",
		},
		datagen.AsExecutor,
	))
	provider.Add("check_value", datagen.NewData(
		[]string{"[repo] check $1 -> $2"},
		[]string{"[repo] $1 -> $2"},
		map[string]string{
			keyKey:   "$1",
			keyValue: "$2",
		},
		datagen.AsChecker,
	))
	providerFail.Add("get_value_not_exists", datagen.NewData(
		[]string{"[repo] get $1"},
		[]string{"[repo] $1 ->"},
		map[string]string{
			keyKey:  "$1",
			keyDist: "$2",
		},
		datagen.AsChecker,
	))
	providerFail.Add("check_value_not_exists", datagen.NewData(
		[]string{"[repo] check $1 -> $2"},
		[]string{"[repo] $1 !-> $2"},
		map[string]string{
			keyKey:   "$1",
			keyValue: "$2",
		},
		datagen.AsChecker,
	))
	gen.AddProvider(name, provider)
	gen.AddProvider(nameFail, providerFail)
}

func (s *addon) OnRegisterExecutor(reg handler.Executor[handler.ExecuteHandler]) {
	reg.Handlers().Register(name, func(parameters handler.Parameters) {
		var (
			key, _        = parameters.Get(keyKey)
			value, _      = parameters.Get(keyValue)
			expiration, _ = parameters.Get(keyExpiration)
		)
		if testing {
			log.Default().Info("[repo] set", key, value)
			return
		}
		_ = s.repo.Set(key, value, expiration)
	})
}

func (s *addon) OnRegisterChecker(reg handler.Executor[handler.CheckHandler]) {
	reg.Handlers().Register(name, func(parameters handler.Parameters) bool {
		var (
			key, _ = parameters.Get(keyKey)
			val, _ = parameters.Get(keyValue)
		)
		valueMap, err := s.repo.HGet(key)
		if err != nil {
			return false
		}
		_, ok := valueMap[val]
		return ok
	})
	reg.Handlers().Register(nameFail, func(parameters handler.Parameters) bool {
		var (
			key, _  = parameters.Get(keyKey)
			val, _  = parameters.Get(keyValue)
			dist, _ = parameters.Get(keyDist)
		)
		//dist!="": check if the map is empty
		//val!="": check if the value does not exist
		valueMap, _ := s.repo.HGet(key)
		if len(valueMap) == 0 {
			return true
		}
		_, ok := valueMap[val]
		return dist == "" && !ok
	})
}

func (s *addon) OnRegisterExtractor(reg handler.Executor[handler.ExtractorHandler]) {
	reg.Handlers().Register(name, func(parameters handler.Parameters) variable.ValueMap {
		var (
			key, _ = parameters.Get(keyKey)
		)
		valueMap, err := s.repo.HGet(key)
		if err != nil {
			return nil
		}
		return variable.NewValueMap(valueMap)
	})
}
func New(repo storage.RedisRepository) plugin.Interface {
	return &addon{
		repo: repo,
	}
}
