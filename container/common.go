package container

import (
	"fmt"
	"goldenglow/node"
	"goldenglow/pkg/log"
	"goldenglow/variable"
)

// 处理单个变量匹配项
func variableGen(
	index int,
	dMatch []string,
	tMatches [][]string,
	T node.Item,
	targetVars variable.Set,
) error {
	if len(dMatch) < 2 {
		return nil
	}

	// 目标变量名
	token := dMatch[1]
	if !variable.Is(token) {
		return nil
	}

	// 越界检查
	if index >= len(tMatches) || len(tMatches[index]) < 2 {
		return fmt.Errorf("invalid variable at index %d", index)
	}

	// 源变量名
	varKey := tMatches[index][1]
	prevVar := T.Variables()[varKey]
	if prevVar == nil {
		return log.NotFound("variable:" + varKey)
	}

	// 生成新变量
	targetVars[token] = variable.New(token, prevVar.Value())
	return nil
}
