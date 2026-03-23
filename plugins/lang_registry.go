package plugins

import (
	"errors"
	"fmt"
)

type langRegistry struct {
	langData map[string]LangGroup
	repo     LangRepo
}

func (l *langRegistry) Register(name string, item LangGroup) error {
	if name == "" {
		return errors.New("empty name")
	}
	if item == nil {
		return errors.New("nil item")
	}
	l.langData[name] = item
	return nil
}

func (l *langRegistry) Init() error {
	for name, group := range l.langData {
		for _, item := range group {
			var (
				tv, rv = item.Get()
			)
			err := l.repo.Save(tv, rv)
			if err != nil {
				return fmt.Errorf("%s: %w", name, err)
			}
		}
	}
	return nil
}
func NewLangRegistry(repo LangRepo) LangRegistry {
	return &langRegistry{
		langData: make(map[string]LangGroup),
		repo:     repo,
	}
}
