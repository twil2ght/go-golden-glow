// Package templateaddon is a template addon for Package template
// - dynamic: making conflictRules dynamically
// - Localize: cache and load hardcoding conflictRules
// TODO: enhance flexibility by banning(or keeping) the specific containers of a given template
package templateaddon

import (
	"encoding/json"
	"goldenglow/pkg/brainsaver"
	"goldenglow/pkg/database"
	"goldenglow/pkg/node"
	"goldenglow/pkg/node/template"
	"goldenglow/plugin"
	"goldenglow/utils"
	"os"
	"path/filepath"
)

func init() {
	plugin.DefaultManager.Register(name, New(database.DefaultJSONRepo()))
}

var (
	name = "template_addon"
)

type addon struct {
	repo database.Repository
}

func (a *addon) Init() {}

func (a *addon) Shutdown() {}

func (a *addon) OnRegisterConflictRule(mgr template.ConflictManager) {
	nodeValueSet, err := a.repo.HGet(brainsaver.KeyNodeSet)
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
				//log.Default().Info("register conflict rule", "origin", nodeValue, "tplToAvoid", tplToAvoid)
				mgr.Register(nodeValue, func(origin node.Interface, tplSet template.Set) {
					delete(tplSet, tplToAvoid)
				})
			}
		}
	}
	mgr.Register("[repo] C @ [Then] A ->", func(origin node.Interface, tplSet template.Set) {
		delete(tplSet, "C @ A")
	})
	mgr.Register("[repo] C @ [Then] A -> B", func(origin node.Interface, tplSet template.Set) {
		delete(tplSet, "C @ A")
	})
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

func New(repo database.Repository) plugin.Interface {
	return &addon{
		repo: repo,
	}
}
