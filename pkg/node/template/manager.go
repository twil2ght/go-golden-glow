package template

import "goldenglow/pkg/registry"

type ConflictSet[T comparable] map[T]map[T]struct{}

var (
	DefaultConflictManager = registry.New[ConflictSet[string]]()
)
