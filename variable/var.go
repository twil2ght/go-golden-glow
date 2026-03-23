package variable

import (
	"goldenglow/pkg/log"
	"regexp"
)

type Item interface {
	Name() string
	Value() string
	Set(value string) error
	OK() bool
	Copy() Item
}
type Parser func(strWithVariables string, variables Set, strict bool) (string, error)
type Set map[string]Item
type Base struct {
	name  string
	value string
}

var (
	VarReg = regexp.MustCompile(`\$\d+`)
)

func (b *Base) Name() string {
	return b.name
}
func (b *Base) Value() string {
	return b.value
}
func (b *Base) OK() bool {
	return b.value != ""
}

func (b *Base) Set(value string) error {
	if value != "" {
		b.value = value
		return nil
	}
	return log.EmptyStrErr()
}

func (b *Base) Copy() Item {
	return New(b.Name(), b.Value())
}

func New(k, v string) Item {
	if k == "" {
		return nil
	}
	if v == "" {
		return nil
	}
	return &Base{
		name:  k,
		value: v,
	}
}
func Copy(target Set) Set {
	dist := make(Set, len(target))
	for k, v := range target {
		dist[k] = New(k, v.Value())
	}
	return dist
}

func ToRawText(target string, variables Set, strict bool) (string, error) {
	var (
		changed  = true
		res      = target
		prev     = target
		shutdown = false
		source   = ""
	)
	for changed {
		changed = false
		prev = res
		res = VarReg.ReplaceAllStringFunc(target, func(s string) string {
			if varb, ok := variables[s]; ok {
				return varb.Value()
			}
			if strict {
				shutdown = true
				source = s
			}
			return s
		})
		if shutdown {
			return "", log.NotExist("variable", source)
		}
		if res != prev {
			changed = true
		}
	}
	return res, nil
}
