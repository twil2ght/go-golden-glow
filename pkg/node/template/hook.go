package template

import "goldenglow/pkg/registry"

type Hook interface {
	OnRegisterConflictSet(mgr registry.Interface[ConflictSet[string]])
}
