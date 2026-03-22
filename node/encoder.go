package node

import (
	"fmt"
	"goldenglow/variable"
)

type templateEncoder struct{}

func DefaultEncoder() Encoder {
	return &templateEncoder{}
}
func (e *templateEncoder) Match(a, b string) bool {
	return e.Do(a) == e.Do(b)
}
func (e *templateEncoder) Do(tpl string) string {
	var (
		idx    int
		seen   = make(map[string]string)
		prefix = "[VAR-"
		suffix = "]"
	)

	return variable.VarReg.ReplaceAllStringFunc(tpl, func(rawVar string) string {
		// 命中缓存：同一个变量，返回同一个标记
		if alias, ok := seen[rawVar]; ok {
			return alias
		}
		// 未命中：生成新标记并缓存
		alias := fmt.Sprintf("%s%d%s", prefix, idx, suffix)
		seen[rawVar] = alias
		idx++
		return alias
	})
}
