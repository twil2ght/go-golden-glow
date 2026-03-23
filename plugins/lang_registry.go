package plugins

import (
	"errors"
)

type langRegistry struct {
	pluginNameSet []string
	repo          LangRepo
}

func (l *langRegistry) Register(name string) error {
	if name == "" {
		return errors.New("empty name")
	}
	l.pluginNameSet = append(l.pluginNameSet, name)
	return nil
}

func (l *langRegistry) RunAll() error {
	return nil
}
func NewLangRegistry(repo LangRepo) LangRegistry {
	return &langRegistry{
		repo: repo,
	}
}
