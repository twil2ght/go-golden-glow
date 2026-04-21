// Package templateaddon is a template addon for Package template
// - dynamic: making conflictRules dynamically
// - Localize: cache and load hardcoding conflictRules
package templateaddon

import (
	"encoding/json"
	"goldenglow/pkg/node"
	"goldenglow/pkg/node/template"
	"goldenglow/pkg/repo"
	"goldenglow/plugin"
	"goldenglow/storage"
	"goldenglow/utils"
	"os"
	"path/filepath"
)

type addon struct {
	repo storage.Repository
}

func (a *addon) Init() {}

func (a *addon) Shutdown() {}

func (a *addon) OnRegisterConflictRule(mgr template.ConflictManager) {
	nodeValueSet, err := a.repo.HGet(repo.KeyNodeSet)
	if err != nil {
		return
	}
	if nodeValueSet == nil {
		return
	}

	for nodeValue := range nodeValueSet {
		for _, tplToAvoid := range loadAllTplToAvoid() {
			if tplToAvoid == nodeValue {
				continue
			}
			if ok, _ := template.MatchTemplate(nodeValue, tplToAvoid); ok {
				tplToAvoid := tplToAvoid
				mgr.Register(nodeValue, func(original, tpl node.Interface) bool {
					return tpl.Value() == tplToAvoid
				})
			}
		}
	}
}
func loadAllTplToAvoid() []string {
	var tplToAvoid []string
	content, _ := os.ReadFile(TplListPath)
	_ = json.Unmarshal(content, &tplToAvoid)
	return tplToAvoid
}

var (
	TplListPath = filepath.Join(utils.RootDir, "config", "tpl_to_avoid.json")
)

func New(repo storage.Repository) plugin.Interface {
	return &addon{
		repo: repo,
	}
}
