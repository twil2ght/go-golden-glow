package calculator

import (
	"fmt"
	"goldenglow/dataGen"
	"goldenglow/executor"
	"goldenglow/executor/checker"
	"goldenglow/executor/extractor"
	"goldenglow/lang"
	"goldenglow/plugins"
	"goldenglow/variable"
	"strconv"
	"strings"
)

func init() {
	if err := plugins.Subscribe(NewCalculator()); err != nil {
		panic(err)
	}
}

const (
	pluginName = "calculator"

	keyExpression = "expression"
	keyLeft       = "left"
	keyRight      = "right"
	keyOperator   = "operator"
)

type calculator struct {
	plugins.Base
}

func (c *calculator) Name() string {
	return pluginName
}

func (c *calculator) OnRegisterExecutor(_ executor.Registry) error {
	return nil
}

func (c *calculator) OnRegisterChecker(reg checker.Registry) error {
	return reg.Register(c.Name(), func(params executor.Parameters) bool {
		if err := executor.Validate(params, keyLeft, keyRight, keyOperator); err != nil {
			return false
		}

		left, err := strconv.ParseFloat(params[keyLeft], 64)
		if err != nil {
			return false
		}

		right, err := strconv.ParseFloat(params[keyRight], 64)
		if err != nil {
			return false
		}

		switch params[keyOperator] {
		case "<", "lt":
			return left < right
		case ">", "gt":
			return left > right
		case "<=", "le":
			return left <= right
		case ">=", "ge":
			return left >= right
		case "=", "==", "eq":
			return left == right
		case "!=", "ne":
			return left != right
		default:
			return false
		}
	})
}

func (c *calculator) OnRegisterExtractor(reg extractor.Registry) error {
	return reg.Register(c.Name(), func(params executor.Parameters) (variable.Item, error) {
		if err := executor.Validate(params, keyExpression); err != nil {
			return nil, err
		}

		expression := params[keyExpression]
		result, err := c.calculate(expression)
		if err != nil {
			return nil, fmt.Errorf("calculation failed: %w", err)
		}

		return variable.New(keyExpression, result), nil
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
	// Handle + and -
	idx := strings.LastIndexAny(expr, "+-")
	if idx == -1 || idx == 0 {
		return c.parseMulDiv(expr)
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
	// Handle * and /
	idx := strings.LastIndexAny(expr, "*/")
	if idx == -1 || idx == 0 {
		return c.parseNumber(expr)
	}

	left, err := c.parseMulDiv(expr[:idx])
	if err != nil {
		return 0, err
	}

	right, err := c.parseNumber(expr[idx+1:])
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

func (c *calculator) parseNumber(expr string) (float64, error) {
	expr = strings.TrimSpace(expr)

	// Handle parentheses
	if strings.HasPrefix(expr, "(") && strings.HasSuffix(expr, ")") {
		return c.parseAddSub(expr[1 : len(expr)-1])
	}

	return strconv.ParseFloat(expr, 64)
}

func (c *calculator) OnRegisterDataGen(reg dataGen.Registry) error {
	generator := dataGen.NewGenerator(c.Name())
	generator.Add("calc", dataGen.New(
		[]string{
			"calculate $1",
			"what is $1",
			"compute $1",
			"$1 equals",
		},
		dataGen.Parameters{
			keyExpression: "$1",
		},
		dataGen.LangTypeDefault,
	))
	generator.Add("compare_lt", dataGen.New(
		[]string{
			"is $1 less than $2",
			"is $1 smaller than $2",
			"$1 < $2",
		},
		dataGen.Parameters{
			keyLeft:     "$1",
			keyRight:    "$2",
			keyOperator: "<",
		},
		dataGen.LangTypeCheck,
	))
	generator.Add("compare_gt", dataGen.New(
		[]string{
			"is $1 greater than $2",
			"is $1 larger than $2",
			"$1 > $2",
		},
		dataGen.Parameters{
			keyLeft:     "$1",
			keyRight:    "$2",
			keyOperator: ">",
		},
		dataGen.LangTypeCheck,
	))
	generator.Add("compare_eq", dataGen.New(
		[]string{
			"is $1 equal to $2",
			"does $1 equal $2",
			"$1 = $2",
			"$1 == $2",
		},
		dataGen.Parameters{
			keyLeft:     "$1",
			keyRight:    "$2",
			keyOperator: "=",
		},
		dataGen.LangTypeCheck,
	))
	return reg.AddGenerator(c.Name(), generator)
}

func (c *calculator) OnRegisterLang(reg lang.Registry) error {
	return reg.Register(c.Name())
}

func NewCalculator() plugins.Item {
	return &calculator{}
}
