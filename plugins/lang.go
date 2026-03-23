package plugins

import (
	"goldenglow/m"
)

type language struct {
	tv, rv m.Hash
}

func (l *language) Get() (tv, rv m.Hash) {
	return l.tv, l.rv
}
