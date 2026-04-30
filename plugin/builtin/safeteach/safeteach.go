package safeteach

import (
	"goldenglow/pkg/datagen"
	"goldenglow/pkg/log"
	"goldenglow/pkg/node/handler"
	"goldenglow/pkg/registry"
	"goldenglow/pkg/runner"
	"goldenglow/plugin"
	"goldenglow/utils"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func init() {
	plugin.DefaultManager.Register(name, &safeTeach{registry.New[bool]()})
}

var (
	name = "safeTeach"

	keyKey   = "key"
	keyValue = "value"
)

type safeTeach struct {
	cache registry.Interface[bool]
}

func (s *safeTeach) OnRegisterIdleHandler(mgr registry.Interface[runner.IdleHandler]) {
	mgr.Register(name, func() bool {
		s.Reset()
		return true
	})
}
func (s *safeTeach) OnRegisterExecutor(reg handler.Executor[handler.ExecuteHandler]) {
	reg.Handlers().Register(name, func(parameters handler.Parameters) {
		var (
			key, _   = parameters.Get(keyKey)
			value, _ = parameters.Get(keyValue)
			cfg, _   = ReadConfig()
		)
		cfg[key] = value == "on"
		s.cache.Unregister(key)
		WriteConfig(cfg)
	})
}

func (s *safeTeach) OnRegisterChecker(reg handler.Executor[handler.CheckHandler]) {
	reg.Handlers().Register(name, func(parameters handler.Parameters) bool {
		var (
			cfg, _ = ReadConfig()
			key, _ = parameters.Get(keyKey)
		)
		log.Default().Debug("[safeTeach] checking", "key", key)
		if cache, _ := s.cache.Get(key); cache {
			return false
		}
		s.cache.Register(key, true)
		return cfg[key]
	})
}

func (s *safeTeach) OnRegisterDataGen(gen datagen.Generator) {
	provider := datagen.NewProvider()
	provider.Add("set", datagen.NewData(
		[]string{"[ST] $1 -> $2"},
		[]string{},
		map[string]string{
			keyKey:   "$1",
			keyValue: "$2",
		},
		datagen.AsExecutor,
	))
	provider.Add("teach", datagen.NewData(
		[]string{"check [teach]"},
		[]string{"Zero starts to teach Susie"},
		map[string]string{
			keyKey: "teach",
		},
		datagen.AsChecker,
	))
	provider.Add("ask", datagen.NewData(
		[]string{"check [ask]"},
		[]string{"Susie is enabled to ask Zero questions"},
		map[string]string{
			keyKey: "ask",
		},
		datagen.AsChecker,
	))
	gen.AddProvider(name, provider)
}

func (s *safeTeach) Init() {}

func (s *safeTeach) Shutdown() {}
func (s *safeTeach) Reset()    { s.cache = registry.New[bool]() }

type Config map[string]bool

var (
	path = filepath.Join(utils.RootDir, "config/safeTeach.yaml")
)

func ReadConfig() (mode Config, ok bool) {
	content, err := os.ReadFile(path)
	if err != nil {
		return Config{}, false
	}

	var m Config
	err = yaml.Unmarshal(content, &m)
	if err != nil {
		return Config{}, false
	}
	return m, true
}
func WriteConfig(config Config) {
	content, err := yaml.Marshal(config)
	if err != nil {
		return
	}
	err = os.WriteFile(path, content, 0644)
	if err != nil {
		return
	}
}
