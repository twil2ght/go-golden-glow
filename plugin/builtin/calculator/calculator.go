package calculator

import (
	"fmt"
	"goldenglow/m"
	"goldenglow/pkg/datagen"
	"goldenglow/pkg/node/handler"
	"goldenglow/pkg/variable"
	"goldenglow/plugin"
	"strconv"
	"strings"
)

func init() {
	plugin.DefaultManager.Register(name, NewCalculator())
}

const (
	name = "calculator"

	keyExpression = "expression"
	keyLeft       = "left"
	keyRight      = "right"
	keyOperator   = "operator"
	keyDist       = "dist"
)

type calculator struct{}

func (c *calculator) OnRegisterChecker(reg handler.Executor[handler.CheckHandler]) {
	reg.Handlers().Register(name, func(parameters handler.Parameters) bool {
		left, err := parameters.Get(keyLeft)
		if err != nil {
			return false
		}
		right, err := parameters.Get(keyRight)
		if err != nil {
			return false
		}
		operator, err := parameters.Get(keyOperator)
		if err != nil {
			return false
		}

		leftVal, err := strconv.ParseFloat(left, 64)
		if err != nil {
			return false
		}

		rightVal, err := strconv.ParseFloat(right, 64)
		if err != nil {
			return false
		}

		switch operator {
		case "<", "lt":
			return leftVal < rightVal
		case ">", "gt":
			return leftVal > rightVal
		case "<=", "le":
			return leftVal <= rightVal
		case ">=", "ge":
			return leftVal >= rightVal
		case "=", "==", "eq":
			return leftVal == rightVal
		case "!=", "ne":
			return leftVal != rightVal
		default:
			return false
		}
	})
}

func (c *calculator) OnRegisterExtractor(reg handler.Executor[handler.ExtractorHandler]) {
	reg.Handlers().Register(name, func(parameters handler.Parameters) variable.ValueMap {
		expression, err := parameters.Get(keyExpression)
		if err != nil {
			return nil
		}

		result, err := c.calculate(expression)
		if err != nil {
			return nil
		}

		return variable.NewValueMap(m.Hash{result: struct{}{}})
	})
}

func (c *calculator) calculate(expression string) (string, error) {
	expression = strings.TrimSpace(expression)
	expression = strings.ReplaceAll(expression, " ", "")

	// Support basic operators: +, -, *, /, (, )
	result, err := c.evalExpression(expression)
	if err != nil {
		return "", err
	}

	return strconv.FormatFloat(result, 'f', -1, 64), nil
}

func (c *calculator) evalExpression(expr string) (float64, error) {
	// Simple recursive descent parser for basic arithmetic
	return c.parseAddSub(expr)
}

func (c *calculator) parseAddSub(expr string) (float64, error) {
	idx := lastBinaryOpOutsideParens(expr, "+-")
	if idx == -1 {
		return c.parseMulDiv(expr)
	}
	if idx == 0 {
		return c.parseUnary(expr)
	}

	left, err := c.parseAddSub(expr[:idx])
	if err != nil {
		return 0, err
	}

	right, err := c.parseMulDiv(expr[idx+1:])
	if err != nil {
		return 0, err
	}

	switch expr[idx] {
	case '+':
		return left + right, nil
	case '-':
		return left - right, nil
	}

	return 0, fmt.Errorf("invalid operator")
}

func (c *calculator) parseMulDiv(expr string) (float64, error) {
	idx := lastBinaryOpOutsideParens(expr, "*/")
	if idx == -1 || idx == 0 {
		return c.parseUnary(expr)
	}

	left, err := c.parseMulDiv(expr[:idx])
	if err != nil {
		return 0, err
	}

	right, err := c.parseUnary(expr[idx+1:])
	if err != nil {
		return 0, err
	}

	switch expr[idx] {
	case '*':
		return left * right, nil
	case '/':
		if right == 0 {
			return 0, fmt.Errorf("division by zero")
		}
		return left / right, nil
	}

	return 0, fmt.Errorf("invalid operator")
}

func (c *calculator) parseUnary(expr string) (float64, error) {
	if strings.HasPrefix(expr, "-") {
		val, err := c.parseUnary(expr[1:])
		if err != nil {
			return 0, err
		}
		return -val, nil
	}
	if strings.HasPrefix(expr, "+") {
		return c.parseUnary(expr[1:])
	}
	return c.parseNumber(expr)
}

func (c *calculator) parseNumber(expr string) (float64, error) {
	expr = strings.TrimSpace(expr)

	if isWrappedByParens(expr) {
		return c.parseAddSub(expr[1 : len(expr)-1])
	}

	return strconv.ParseFloat(expr, 64)
}

func isWrappedByParens(expr string) bool {
	if len(expr) < 2 || expr[0] != '(' || expr[len(expr)-1] != ')' {
		return false
	}
	depth := 0
	for i, ch := range expr {
		if ch == '(' {
			depth++
		} else if ch == ')' {
			depth--
		}
		if depth == 0 && i < len(expr)-1 {
			return false
		}
	}
	return depth == 0
}

// lastBinaryOpOutsideParens finds the last binary operator (not unary) outside parentheses.
// A + or - at position 0 or preceded by an operator/open-paren is treated as unary and skipped.
func lastBinaryOpOutsideParens(s, charset string) int {
	depth := 0
	for i := len(s) - 1; i >= 0; i-- {
		switch s[i] {
		case ')':
			depth++
		case '(':
			depth--
		default:
			if depth == 0 && strings.ContainsRune(charset, rune(s[i])) {
				if strings.ContainsRune("+-", rune(s[i])) {
					if i == 0 || strings.ContainsRune("+-*/(", rune(s[i-1])) {
						continue
					}
				}
				return i
			}
		}
	}
	return -1
}

// OnRegisterDataGen registers data generation for calculator
func (c *calculator) OnRegisterDataGen(gen datagen.Generator) {
	provider := datagen.NewProvider()
	provider.Add("compare_lt", datagen.NewData(
		[]string{"[compute] check $1 < $2"},
		[]string{"[compute] $1 < $2"},
		map[string]string{
			keyLeft:     "$1",
			keyRight:    "$2",
			keyOperator: "<",
		},
		datagen.AsChecker,
	))
	provider.Add("compare_gt", datagen.NewData(
		[]string{"[compute] check $1 > $2"},
		[]string{"[compute] $1 > $2"},
		map[string]string{
			keyLeft:     "$1",
			keyRight:    "$2",
			keyOperator: ">",
		},
		datagen.AsChecker,
	))
	provider.Add("compare_eq", datagen.NewData(
		[]string{"[compute] check $1 = $2"},
		[]string{"[compute] $1 = $2"},
		map[string]string{
			keyLeft:     "$1",
			keyRight:    "$2",
			keyOperator: "=",
		},
		datagen.AsChecker,
	))
	provider.Add("plus", datagen.NewData(
		[]string{"[compute] get $1 + $2"},
		[]string{"[compute] $1 + $2 -> $3"},
		map[string]string{
			keyExpression: "$1+$2",
			keyDist:       "$3",
		},
		datagen.AsExtractor,
	))
	provider.Add("sub", datagen.NewData(
		[]string{"[compute] get $1 - $2"},
		[]string{"[compute] $1 - $2 -> $3"},
		map[string]string{
			keyExpression: "$1-$2",
			keyDist:       "$3",
		},
		datagen.AsExtractor,
	))
	provider.Add("multiply", datagen.NewData(
		[]string{"[compute] get $1 * $2"},
		[]string{"[compute] $1 * $2 -> $3"},
		map[string]string{
			keyExpression: "$1*$2",
			keyDist:       "$3",
		},
		datagen.AsExtractor,
	))
	gen.AddProvider(name, provider)
}

func (c *calculator) Init() {}

func (c *calculator) Shutdown() {}

func NewCalculator() plugin.Interface {
	return &calculator{}
}
