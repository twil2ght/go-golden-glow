package node

import (
	"fmt"
)

type templateEncoder struct {
	varReplacer VarReplacer
}

func NewEncoder(varReplacer VarReplacer) (Encoder, error) {
	if varReplacer == nil {
		return nil, fmt.Errorf("encoder init:varReplacer is nil")
	}
	return &templateEncoder{
		varReplacer: varReplacer,
	}, nil
}
func (e *templateEncoder) Match(a, b string) bool {
	return e.Do(a) == e.Do(b)
}
func (e *templateEncoder) Do(tpl string) string {
	var (
		idx  int
		seen = make(map[string]string)
	)

	return e.varReplacer.ReplaceAllStringFunc(tpl, func(rawVar string) string {

		if alias, ok := seen[rawVar]; ok {
			return alias
		}

		alias := fmt.Sprintf("$%d", idx)
		seen[rawVar] = alias
		idx++
		return alias
	})
}
