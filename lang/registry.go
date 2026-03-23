package lang

import "errors"

type langRegistry struct {
	pluginNameSet []string
	repo          Repo
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
func NewLangRegistry(repo Repo) Registry {
	return &langRegistry{
		repo: repo,
	}
}
