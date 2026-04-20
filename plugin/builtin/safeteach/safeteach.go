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
)

type safeTeach struct{}

func (s *safeTeach) OnCheckExecutor(reg handler.Executor[handler.CheckHandler]) {
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
	gen.AddProvider(name, provider)
}

func (s *safeTeach) Init() {}

func (s *safeTeach) Shutdown() {}
func ReadConfig() (ok bool, mode bool) {
	jsonFile := utils.FindAllJsonFiles(filepath.Join(utils.RootDir, "config/safeTeach.json"))
	if len(jsonFile) == 0 {
		return false, false
	}
	content, err := os.ReadFile(jsonFile[0])
	if err != nil {
		return false, false
	}
	type Mode struct {
		Teach bool `json:"teach"`
	}
	var m Mode
	err = json.Unmarshal(content, &m)
	if err != nil {
		return false, false
	}
	return true, m.Teach
}
