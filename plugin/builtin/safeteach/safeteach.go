package safeteach

import (
	"encoding/json"
	"goldenglow/pkg/datagen"
	"goldenglow/pkg/node/handler"
	"goldenglow/plugin"
	"goldenglow/utils"
	"os"
	"path/filepath"
)

func init() {
	plugin.DefaultManager.Register(name, &safeTeach{})
}

var (
	name = "safeTeach"

	keyMode = "mode"
)

type safeTeach struct{}

func (s *safeTeach) OnRegisterExecutor(reg handler.Executor[handler.ExecuteHandler]) {
	reg.Handlers().Register(name, func(parameters handler.Parameters) {
		mode, _ := parameters.Get(keyMode)
		WriteConfig(mode == "on")
	})
}

func (s *safeTeach) OnRegisterChecker(reg handler.Executor[handler.CheckHandler]) {
	reg.Handlers().Register(name, func(parameters handler.Parameters) bool {
		_, mode := ReadConfig()
		return mode
	})
}

func (s *safeTeach) OnRegisterDataGen(gen datagen.Generator) {
	provider := datagen.NewProvider()
	provider.Add("mode", datagen.NewData(
		[]string{"check [mode]"},
		[]string{"Zero starts to teach Susie"},
		map[string]string{},
		datagen.AsChecker,
	))
	provider.Add("set", datagen.NewData(
		[]string{"[mode] $1"},
		[]string{},
		map[string]string{
			keyMode: "$1",
		},
		datagen.AsExecutor,
	))
	gen.AddProvider(name, provider)
}

func (s *safeTeach) Init() {}

func (s *safeTeach) Shutdown() {}

type Mode struct {
	Teach bool `json:"teach"`
}

var (
	path = filepath.Join(utils.RootDir, "config/safeTeach.json")
)

func ReadConfig() (ok bool, mode bool) {
	content, err := os.ReadFile(path)
	if err != nil {
		return false, false
	}

	var m Mode
	err = json.Unmarshal(content, &m)
	if err != nil {
		return false, false
	}
	return true, m.Teach
}
func WriteConfig(mode bool) {
	var m = Mode{Teach: mode}
	content, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return
	}
	err = os.WriteFile(path, content, 0644)
	if err != nil {
		return
	}
}
