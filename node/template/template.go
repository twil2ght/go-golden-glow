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

// Source TODO wl,bl,common
type Source interface {
	Get() (node.Set, error)
	Set(key, value string) error
	Common() node.Set
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
	templates, err := c.source.Get()
	if err != nil {
		return nil, err
	}
	set := c.toSpecific(tar, templates)
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
	if template == "" || len(strings.Fields(target)) < len(strings.Fields(template)) {
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
			parts = append(parts, `(.+)`)
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
	//TODO var需要传value而不是head
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

func (c *core) toSpecific(target node.Item, templates node.Set) node.Set {
	matches := make(node.Set)

	for key, n := range templates {
		if ok, vars := c.matchTemplate(target.Value(), n.Value()); ok {
			err := n.SetVariable(vars)
			if err != nil {
				continue
			}
			matches[key] = n
		}
	}

	if len(matches) == 0 {
		return nil
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

	return result
}
func (c *core) Match(a, b string) bool {
	ok, _ := c.matchTemplate(a, b)
	return ok
}
