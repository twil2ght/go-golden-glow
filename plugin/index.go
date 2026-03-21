package plugin

import (
	"fmt"
	"goldenglow/container"
	"goldenglow/m"
	"goldenglow/node"
	"goldenglow/pkg/log"
	"goldenglow/utils"
	"strings"
)

type Handler func(ctx HandlerCtx)
type TRGroup map[string]TR
type NV []string
type ParamCfg struct {
	Head        string
	ParamLength int
}
type TR struct {
	TV NV
	RV NV
}
type LangRegistry interface {
	Register(name string, langAPI TRGroup) error
}
type baseLangRegistry struct {
	LangAPI map[string]TRGroup
	store   container.Store
}
type Base struct {
	ParamCfg    *ParamCfg
	NodeCreator node.Creator
	LangAPI     TRGroup
	Name        string
}
type Context interface {
	Get() m.H
	Set(key string, component any)
}
type HandlerCtxBase struct {
	data m.H
}

func (c *HandlerCtxBase) Get() m.H {
	return c.data
}
func (c *HandlerCtxBase) Set(params m.H) {
	c.data = params
}

type HandlerCtx interface {
	Set(params m.H)
	Get() m.H
}
type Item interface {
	Register(c Context) error
}

const (
	Event            = "event"
	LangRegistryName = "LangRegistry"
	NodeRegistry     = "NodeRegistry"
)

func (lc *baseLangRegistry) Register(name string, group TRGroup) error {
	head := "lang registry"
	if name == "" {
		return fmt.Errorf("%s name is empty", head)
	}
	if group == nil {
		return fmt.Errorf("%s group is nil", head)
	}
	lc.LangAPI[name] = group
	return nil
}

func (lc *baseLangRegistry) Init() error {
	head := "lang registry"
	for _, group := range lc.LangAPI {
		for _, rule := range group {
			err := lc.store.Do(m.ToHash(rule.TV), m.ToHash(rule.RV))
			if err != nil {
				return fmt.Errorf("%s: %s", head, err)
			}
		}
	}
	return nil
}
func NewLangRegistry(store container.Store) (LangRegistry, error) {
	if store == nil {
		return nil, fmt.Errorf("new lang registry: store is nil")
	}
	return &baseLangRegistry{
		LangAPI: make(map[string]TRGroup),
		store:   store,
	}, nil
}

// Parse Format: [Head] [Parameters]
// By default,the last parameter should be userID
// requiredLength==-1 -> indefinite
func Parse(raw string, requiredHead string, requiredLen int) ([]string, bool, error) {
	fields := strings.Fields(raw)
	if len(fields) == 0 {
		return nil, false, log.LengthErr()
	}
	head := fields[0]
	if head != requiredHead {
		return nil, false, fmt.Errorf("wrong head:%s", raw)
	}
	paramsStr := strings.Join(fields[1:], " ")
	params := utils.TrimAll(strings.Split(paramsStr, "&"))
	if len(params) != requiredLen && requiredLen != -1 {
		return nil, false, fmt.Errorf("paramLength-get:%d-required:%d", len(params), requiredLen)
	}
	return params, true, nil
}
func ParseFunc(n node.Base, ParamConfig *ParamCfg, ParamKeys map[string]int) (map[string]any, error) {
	nv, err := n.ToText()
	if err != nil {
		return nil, fmt.Errorf("%s Decode fail", ParamConfig.Head)
	}
	parameters, ok, err := Parse(nv, ParamConfig.Head, ParamConfig.ParamLength)
	if !ok {
		return nil, err
	}

	result := make(map[string]any)
	paramLen := len(parameters)

	var (
		warning    error
		warningMsg = "paramKeys don't match params"
	)
	for k, idx := range ParamKeys {
		if idx == -1 {
			fixedCount := len(ParamKeys) - 1
			if fixedCount >= 0 && fixedCount <= paramLen {
				result[k] = parameters[fixedCount:]
			} else {
				warning = fmt.Errorf("%s", warningMsg)
				result[k] = []any{}
			}
			continue
		}

		if idx >= 0 && idx < paramLen {
			result[k] = parameters[idx]
		} else {
			warning = fmt.Errorf("%s", warningMsg)
			result[k] = ""
		}
	}

	return result, warning
}

// Execute 核心执行函数：根据事件分发处理逻辑，并返回标准化日志
// 参数说明：
//   - Head: 日志头部标识（如"[BetterAskNode]"）
//   - params: 解析后的参数映射
//   - paramKeys: 参数key与索引的映射（预留扩展，比如参数校验）
//   - c: 处理器上下文
//   - EventFunc: 事件与处理器的映射表
//
// 返回值：标准化的日志对象
func Execute(
	Head string,
	parser func() (m.H, error),
	c HandlerCtx,
	EventFunc map[string]Handler) error {
	errStr := ""
	errMsg := Head + " [Execute] " + errStr

	params, err := parser()
	if err != nil {
		errStr = err.Error()
		return fmt.Errorf("%s", errMsg)
	}
	if c == nil {
		errStr = "handler context is nil"
		return fmt.Errorf("%s", errMsg)
	}
	if len(EventFunc) == 0 {
		errStr = "handlers map is empty"
		return fmt.Errorf("%s", errMsg)
	}

	c.Set(params)

	eventVal, ok := params[Event]
	if !ok {
		errStr = "param:event not found"
		return fmt.Errorf("%s", errMsg)
	}
	event, ok := eventVal.(string)
	if !ok {
		errStr = "param: event not string type"
		return fmt.Errorf("%s", errMsg)
	}
	handler, handlerOk := EventFunc[event]
	if handlerOk {
		defer func() {
			if r := recover(); r != nil {
				errStr = "event handler panic: " + fmt.Sprintf("%v", r)
			}
		}()
		handler(c)
		return nil
	}
	errStr = fmt.Sprintf("no handler found for event [%s]", event)
	return fmt.Errorf("%s", errMsg)

}

func RegisterConfig(
	ctx Context,
	paramCfg *ParamCfg,
	nodeCreator node.Creator,
	langAPI TRGroup,
	pluginName string,
) error {
	if paramCfg == nil {
		return fmt.Errorf("parameter config is nil")
	}
	if nodeCreator == nil {
		return fmt.Errorf("nodeCreator is nil")
	}
	if pluginName == "" {
		return fmt.Errorf("pluginName is empty")
	}
	if langAPI == nil {
		return fmt.Errorf("langAPI is nil")
	}

	langCore, ok := ctx.Get()[LangRegistryName].(LangRegistry)
	if !ok {
		return fmt.Errorf("couldn't find LangRegistry in context")
	}
	err := langCore.Register(pluginName, langAPI)
	if err != nil {
		return err
	}

	nodeRegistry, ok := ctx.Get()[NodeRegistry].(node.Registry)
	if !ok {
		return fmt.Errorf("couldn't find nodeRegistry in context")
	}
	err = nodeRegistry.Register(paramCfg.Head, nodeCreator)
	if err != nil {
		return err
	}

	return nil
}

func (p *Base) Register(c Context) error {
	return RegisterConfig(
		c,
		p.ParamCfg,
		p.NodeCreator,
		p.LangAPI,
		p.Name,
	)
}
