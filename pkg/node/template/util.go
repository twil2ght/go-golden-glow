package template

import (
	"goldenglow/pkg/variable"
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
func MatchTemplate(target, template string) (bool, variable.Set) {
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

const (
	maxSolutions        = 3
	maxCandidatesPerVar = 10
)

func MatchTemplateAll(target, template string) []variable.Set {
	if template == "" {
		return nil
	}
	segments := segment(template)
	if len(segments) == 0 {
		if target == template {
			return []variable.Set{{}}
		}
		return nil
	}

	var result []variable.Set
	resPtr := &result

	var backtrack func(pos int, segIdx int, bind variable.Set)
	backtrack = func(pos int, segIdx int, bind variable.Set) {
		if len(*resPtr) >= maxSolutions {
			return
		}
		if segIdx >= len(segments) {
			if pos == len(target) {
				*resPtr = append(*resPtr, bind.Clone())
			}
			return
		}

		seg := segments[segIdx]
		if variable.VarReg.MatchString(seg) {
			var stopLit string
			nextSegIdx := segIdx + 1
			if nextSegIdx < len(segments) {
				stopLit = segments[nextSegIdx]
			}

			rem := target[pos:]
			var takeCandidates []int

			if stopLit == "" {
				if len(rem) >= 1 {
					takeCandidates = append(takeCandidates, len(rem))
				}
			} else {
				searchStart := 0
				for {
					off := strings.Index(rem[searchStart:], stopLit)
					if off == -1 {
						break
					}
					absInRem := searchStart + off
					if absInRem >= 1 {
						takeCandidates = append(takeCandidates, absInRem)
						if len(takeCandidates) >= maxCandidatesPerVar {
							break
						}
					}
					searchStart = absInRem + len(stopLit)
				}
			}

			for _, takeLen := range takeCandidates {
				val := rem[:takeLen]
				newBind := bind.Clone()
				newBind[seg] = variable.New(seg, val)

				nextPos := pos + takeLen + len(stopLit)
				backtrack(nextPos, segIdx+2, newBind)

				if len(*resPtr) >= maxSolutions {
					break
				}
			}
		} else {
			lit := seg
			if !strings.HasPrefix(target[pos:], lit) {
				return
			}
			backtrack(pos+len(lit), segIdx+1, bind)
		}
	}

	backtrack(0, 0, make(variable.Set))
	return *resPtr
}
