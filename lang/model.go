package lang

import "goldenglow/m"

type Repo interface {
	Save(tv, rv m.Hash) error
}
type Registry interface {
	Register(pluginName string) error //传pluginName去找data路径
	RunAll() error
}
