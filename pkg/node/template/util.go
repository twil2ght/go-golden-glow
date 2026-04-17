package template

import (
	"goldenglow/variable"
	"regexp"
	"strings"
)

func segment(tpl string) []string {
	var segments []string
	last := 0

	for _, m := range variable.VarReg.FindAllStringSubmatchIndex(tpl, -1) {
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
func matchTemplate(target, template string) (bool, variable.Set) {
	if template == "" {
		return false, nil
	}

	segments := segment(template)
	if len(segments) == 0 {
		return target == template, nil
	}

	var parts []string
	var keys []string

	for _, seg := range segments {
		if variable.VarReg.MatchString(seg) {
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

//func migrateVarSet(varFrom, varTo variable.Set) error {
//	if varTo == nil {
//		return fmt.Errorf("clean variables: varTo is nil")
//	}
//	if varFrom == nil {
//		return fmt.Errorf("clean variables: varFrom is nil")
//	}
//	for key, e := range varTo {
//		val, err := variable.ToRawText(e.Value(), varFrom, false)
//		if err != nil {
//			return fmt.Errorf("clean variables: %w", err)
//		}
//		err = e.Set(val)
//		if err != nil {
//			return fmt.Errorf("clean variables: %w", err)
//		}
//		varTo[key] = e
//	}
//	return nil
//}
