package repoaddon

import (
	"fmt"
	"goldenglow/pkg/database"
	"goldenglow/pkg/datagen"
	"goldenglow/pkg/log"
	"goldenglow/pkg/node/handler"
	"goldenglow/pkg/registry"
	"goldenglow/pkg/runner"
	"goldenglow/pkg/variable"
	"goldenglow/plugin"
)

func init() {
	plugin.DefaultManager.Register(name, New(database.DefaultRedisRepo()))
}

const (
	name     = "repo_addon"
	nameFail = "repo_addon_fail"
	testing  = false
	// parameter keys
	keyKey         = "key"
	keyValue       = "value"
	keyExpiration  = "expiration"
	keyDist        = "dist"
	KeySingleValue = "singleValue"
)

type addon struct {
	repo  database.RedisRepository
	cache registry.Interface[bool]
}

func (s *addon) OnRegisterIdleHandler(mgr registry.Interface[runner.IdleHandler]) {
	mgr.Register(name, func() bool {
		s.Reset()
		return true
	})
}

func (s *addon) Init() {}

func (s *addon) Shutdown() {}
func (s *addon) Reset()    { s.cache = registry.New[bool]() }
func (s *addon) OnRegisterDataGen(gen datagen.Generator) {
	provider := datagen.NewProvider()
	providerFail := datagen.NewProvider()
	provider.Add("get_value", datagen.NewData(
		[]string{"[repo] [GET] $1"},
		[]string{"[repo] $1 -> $2"},
		map[string]string{
			keyKey:  "$1",
			keyDist: "$2",
		},
		datagen.AsExtractor,
	))
	provider.Add("get_value_with_cond", datagen.NewData(
		[]string{"[repo] [GET] $1 @Caller $4"},
		[]string{"[repo] $4 @ $1 -> $2"},
		map[string]string{
			keyKey:  "$1",
			keyDist: "$2",
		},
		datagen.AsExtractor,
	))
	provider.Add("set_value", datagen.NewData(
		[]string{"[repo] [SET] $1 -> $2"},
		[]string{},
		map[string]string{
			keyKey:   "$1",
			keyValue: "$2",
		},
		datagen.AsExecutor,
	))
	provider.Add("set_single_value", datagen.NewData(
		[]string{"[repo] [SSET] $1 -> $2"},
		[]string{},
		map[string]string{
			keyKey:         "$1",
			keyValue:       "$2",
			KeySingleValue: "true",
		},
		datagen.AsExecutor,
	))
	provider.Add("check_value", datagen.NewData(
		[]string{"[repo] [CHECK] $1 -> $2"},
		[]string{"[repo] $1 -> $2"},
		map[string]string{
			keyKey:   "$1",
			keyValue: "$2",
		},
		datagen.AsChecker,
	))
	providerFail.Add("get_value_not_exists_with_cond", datagen.NewData(
		[]string{"[repo] [GET] $1 @Caller $4"},
		[]string{"[repo] $4 @ $1 ->"},
		map[string]string{
			keyKey:  "$1",
			keyDist: "$2",
		},
		datagen.AsChecker,
	))
	providerFail.Add("get_value_not_exists", datagen.NewData(
		[]string{"[repo] [GET] $1"},
		[]string{"[repo] $1 ->"},
		map[string]string{
			keyKey:  "$1",
			keyDist: "$2",
		},
		datagen.AsChecker,
	))
	providerFail.Add("check_value_not_exists", datagen.NewData(
		[]string{"[repo] [CHECK] $1 -> $2"},
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
			key, _         = parameters.Get(keyKey)
			value, _       = parameters.Get(keyValue)
			singleValue, _ = parameters.Get(KeySingleValue)
			expiration, _  = parameters.Get(keyExpiration)
		)
		log.Default().Info("[repo] set", key, value)
		if testing {
			return
		}
		s.cache.Unregister(fmt.Sprintf("check %s!->%s", key, value))
		if singleValue == "true" {
			s.repo.Del(key)
		}
		_ = s.repo.Set(key, value, expiration)
	})
}

func (s *addon) OnRegisterChecker(reg handler.Executor[handler.CheckHandler]) {
	// check $1 -> $2: $1 -> $2($1 !-> $2)
	reg.Handlers().Register(name, func(parameters handler.Parameters) bool {
		var (
			key, _ = parameters.Get(keyKey)
			val, _ = parameters.Get(keyValue)
		)
		//log.Default().Debug("[repo] check", key, val)
		if e, _ := s.cache.Get(fmt.Sprintf("check %s->%s", key, val)); e {
			return false
		}
		s.cache.Register(fmt.Sprintf("check %s->%s", key, val), true)
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
		//log.Default().Debug("[repo] check (fail)", "key", key)
		if e, _ := s.cache.Get(fmt.Sprintf("check %s!->%s", key, val)); e {
			return false
		}
		s.cache.Register(fmt.Sprintf("check %s!->%s", key, val), true)
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
		for val := range valueMap {
			//log.Default().Debug("[repo] get", key, val)
			if e, _ := s.cache.Get(fmt.Sprintf("%s->%s", key, val)); e {
				delete(valueMap, val)
			} else {
				s.cache.Register(fmt.Sprintf("%s->%s", key, val), true)
				s.cache.Register(fmt.Sprintf("check %s->%s", key, val), true)
				s.cache.Register(fmt.Sprintf("check %s!->%s", key, val), true)
			}
		}
		if err != nil {
			return nil
		}
		return variable.NewValueMap(valueMap)
	})
}
func New(repo database.RedisRepository) plugin.Interface {
	return &addon{
		repo:  repo,
		cache: registry.New[bool](),
	}
}
