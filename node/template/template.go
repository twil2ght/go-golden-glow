package template

import (
	"fmt"
	"goldenglow/node"
	"goldenglow/variable"
	"regexp"
	"strings"
)

type Core interface {
	Get(tar node.Item) (node.Set, error)
	Match(input, target string) bool
}

type Source interface {
	GetTemplates() (node.Set, error)
}

type core struct {
	source Source
	varReg *regexp.Regexp
}

func New(source Source, varReg *regexp.Regexp) (Core, error) {
	if source == nil {
		return nil, fmt.Errorf("template core init: source is nil")
	}
	if varReg == nil {
		return nil, fmt.Errorf("template core init: varReg is nil")
	}
	return &core{
		source: source,
		varReg: varReg,
	}, nil
}

func (c *core) Get(tar node.Item) (node.Set, error) {
	templates, err := c.source.GetTemplates()
	if err != nil {
		return nil, err
	}
	//set, err := c.toSpecific(tar, templates)
	set, err := c.AllTemplates(tar, templates)
	if err != nil {
		return nil, err
	}
	return set, nil
}

func (c *core) segment(tpl string) []string {
	var segments []string
	last := 0

	for _, m := range c.varReg.FindAllStringSubmatchIndex(tpl, -1) {
		start, end := m[0], m[1]
		if start > last {
			segments = append(segments, tpl[last:start])
		}
		segments = append(segments, tpl[start:end])
		last = end
	}

	if last < len(tpl) {
		segments = append(segments, tpl[last:])
	}
	return segments
}

func (c *core) matchTemplate(target, template string) (bool, variable.Set) {
	if template == "" {
		return false, nil
	}

	segments := c.segment(template)
	if len(segments) == 0 {
		return target == template, nil
	}

	var parts []string
	var keys []string

	for _, seg := range segments {
		if c.varReg.MatchString(seg) {
			parts = append(parts, `(.+?)`)
			keys = append(keys, seg)
		} else {
			parts = append(parts, regexp.QuoteMeta(seg))
		}
	}

	expr := "^" + strings.Join(parts, "") + "$"
	re, err := regexp.Compile(expr)
	if err != nil {
		return false, nil
	}

	match := re.FindStringSubmatch(target)
	if len(match) != len(keys)+1 {
		return false, nil
	}
	phs := make(variable.Set)
	for i, key := range keys {
		val := match[i+1]
		if val == "" {
			return false, nil
		}
		phs[key] = variable.New(key, val)
	}

	return true, phs
}

func (c *core) toSpecific(target node.Item, templates node.Set) (node.Set, error) {
	matches, err := c.AllTemplates(target, templates)
	if err != nil {
		return nil, err
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("template core to: no template variables found")
	}

	result := make(node.Set)

	for currKey, currNode := range matches {
		keep := true
		replaceKey := ""

		for resKey, resNode := range result {
			// 现有模板能匹配当前 → 当前更具体
			if ok, _ := c.matchTemplate(currNode.Value(), resNode.Value()); ok {
				replaceKey = resKey
				break
			}
			// 当前能匹配现有 → 当前更通用
			if ok, _ := c.matchTemplate(resNode.Value(), currNode.Value()); ok {
				keep = false
				break
			}
		}

		if !keep {
			continue
		}

		if replaceKey != "" {
			result[replaceKey] = currNode
		} else {
			result[currKey] = currNode
		}
	}

	return result, nil
}
func (c *core) Match(a, b string) bool {
	ok, _ := c.matchTemplate(a, b)
	return ok
}
func clean(varFrom, varTo variable.Set) error {
	if varTo == nil {
		return fmt.Errorf("clean variables: varTo is nil")
	}
	if varFrom == nil {
		return fmt.Errorf("clean variables: varFrom is nil")
	}
	for key, e := range varTo {
		val, err := variable.ToRawText(e.Value(), varFrom, false)
		if err != nil {
			return fmt.Errorf("clean variables: %w", err)
		}
		err = e.Set(val)
		if err != nil {
			return fmt.Errorf("clean variables: %w", err)
		}
		varTo[key] = e
	}
	return nil
}
func (c *core) AllTemplates(target node.Item, templates node.Set) (node.Set, error) {
	matches := make(node.Set)

	for key, n := range templates {
		if ok, vars := c.matchTemplate(target.Value(), n.Value()); ok {
			err := clean(target.Variables(), vars)
			if err != nil {
				return nil, err
			}
			err = n.SetVariable(vars)
			if err != nil {
				return nil, err
			}
			matches[key] = n
		}
	}
	var item, ok = target.(*node.Base)
	if !ok {
		return matches, nil
	}

	var raw []string
	for _, hub := range target.VariableSetHub() {
		err := target.SetVariable(hub)
		if err != nil {
			continue
		}
		rawItem, _ := item.ToText()
		raw = append(raw, rawItem)
	}

	for key, n := range templates {
		for _, rawText := range raw {
			if ok, vars := c.matchTemplate(rawText, n.Value()); ok {
				if ok, _ := c.matchTemplate(n.Value(), target.Value()); !ok {
					continue
				}
				err := clean(target.Variables(), vars)
				if err != nil {
					return nil, err
				}
				err = n.SetVariable(vars)
				if err != nil {
					return nil, err
				}
				matches[key] = n
			}
		}
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("no matches found")
	}
	return matches, nil
}
