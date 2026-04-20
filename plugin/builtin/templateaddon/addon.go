// Package templateaddon is a template addon for Package template
// - dynamic: making conflictRules dynamically
// - Localize: cache and load hardcoding conflictRules
package templateaddon

import (
	"goldenglow/pkg/node"
	"goldenglow/pkg/node/template"
	"goldenglow/pkg/repo"
	"goldenglow/storage"
)

type addon struct {
	repo storage.Repository
}

func (a *addon) OnRegisterConflictRule(mgr template.ConflictManager) {
	mgr.Register("", func(original, template node.Interface) bool {
		return false
	})
	nodeValueSet, err := a.repo.HGet(repo.KeyNodeSet)
	if err != nil {
		return
	}
	if nodeValueSet == nil {
		return
	}
	var (
		tplToAvoid = "Zero says if $1 to Susie"
	)
	for nodeValue := range nodeValueSet {
		if ok, _ := template.MatchTemplate(nodeValue, tplToAvoid); ok {
			mgr.Register(nodeValue, func(original, tpl node.Interface) bool {
				return tpl.Value() == tplToAvoid
			})
		}
	}

}
