package plugins

import (
	"fmt"
	"goldenglow/node"
	"strings"
)

type baseNode struct {
	node.Base
	handlers map[string]ExecuteHandler
}

const (
	KeyNamespace = "namespace"
	KeyDefault   = "[node]"
)

func (d *baseNode) Execute() error {
	params := d.GetParams()
	pluginID := params[KeyNamespace]

	if pluginID == "" {
		return fmt.Errorf("pluginID(%s) not found in params", KeyNamespace)
	}

	handler := d.handlers[pluginID]
	if handler == nil {
		return fmt.Errorf("plugin %s not registered", pluginID)
	}

	return handler(params)
}

// GetParams 解析节点文本，提取 [key:value] 格式参数，返回 map[string]string
// 输入示例：[node] [namespace:plugin_name] & [event:Get] & [key:foo]
func (d *baseNode) GetParams() map[string]string {
	// 初始化结果 map
	params := make(map[string]string)

	// 获取节点文本（忽略错误）
	nodeValue, _ := d.ToText()
	if nodeValue == "" {
		return params
	}

	// 1. 按空格分割整个字符串
	parts := strings.Fields(nodeValue)

	// 2. 遍历所有片段，跳过第一个固定的 [node]
	for _, part := range parts[1:] {
		// 3. 清理字符：去掉 [] 和 & 符号
		cleanPart := strings.Trim(part, "[]& ")
		if cleanPart == "" {
			continue
		}

		// 4. 按冒号分割成 key 和 value
		kv := strings.SplitN(cleanPart, ":", 2)
		if len(kv) == 2 {
			key := strings.TrimSpace(kv[0])
			value := strings.TrimSpace(kv[1])
			params[key] = value
		}
	}

	return params
}
