package container

import (
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
	// 目标变量名
	token := dMatch[0]
	// 源变量名
	varKey := tMatches[index][0]
	prevVar := T.Variables()[varKey]
	logger.Debug("prevVar", "", prevVar)
	if prevVar == nil {
		return log.NotFound("variable:" + varKey)
	}

	// 生成新变量
	targetVars[token] = variable.New(token, prevVar.Value())
	return nil
}
