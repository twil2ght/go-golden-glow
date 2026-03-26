package timer

import (
	"fmt"
	"goldenglow/dataGen"
	"goldenglow/executor"
	"goldenglow/executor/checker"
	"goldenglow/executor/extractor"
	"goldenglow/lang"
	"goldenglow/plugins"
	"goldenglow/storage"
	"goldenglow/variable"
	"time"
)

func init() {
	if err := plugins.Subscribe(NewTimer()); err != nil {
		panic(err)
	}
}

const (
	pluginName = "timer"

	//key
	keyMode = "now"
	keyType = "type"
	KeyDist = "dist"

	//mode
	modeFetch  = "fetch"
	modeTicker = "ticker"

	//frequently used types of parameters
	typeTime = "time" //hour:minute
	typeDate = "date" //month:day

	//detailed types of parameters
	typeSecond = "second" //second
	typeHour   = "hour"   //hour
	typeMinute = "minute" //minute
	typeMonth  = "month"  //month
	typeDay    = "day"    //day
	typeYear   = "year"   //year
)

type timer struct {
	plugins.Base
	repo storage.Repository
}

func (t *timer) process(mode, arg string) (string, error) {
	switch mode {
	case modeFetch:
		return t.fetch(arg)
	case modeTicker:
		return "", nil
	default:
		return "", fmt.Errorf("unknown timer mode: %s", mode)
	}
}
func (t *timer) fetch(arg string) (string, error) {
	switch arg {
	case typeTime:
		return t.fetchTime(), nil
	case typeDate:
		return t.fetchDate(), nil
	default:
		return "", fmt.Errorf("unknown timer type: %s", arg)
	}
}
func (t *timer) fetchTime() string {
	var (
		now    = time.Now()
		hour   = now.Hour()
		minute = now.Minute()
		second = now.Second()
	)
	return fmt.Sprintf("%02d:%02d:%02d", hour, minute, second)
}
func (t *timer) fetchDate() string {
	var (
		now   = time.Now()
		month = now.Month()
		day   = now.Day()
	)
	return fmt.Sprintf("%02d:%02d", month, day)
}

func (t *timer) OnRegisterExtractor(reg extractor.Registry) error {
	return reg.Register(pluginName, func(params executor.Parameters) (variable.Item, error) {
		if err := executor.Validate(params, keyMode, keyType, KeyDist); err != nil {
			return nil, err
		}

		value, err := t.process(params[keyMode], params[keyType])
		if err != nil {
			return nil, err
		}

		return variable.New(params[KeyDist], value), err
	})
}

func (t *timer) Name() string {
	return pluginName
}

func (t *timer) OnRegisterChecker(reg checker.Registry) error {
	return reg.Register(pluginName, func(parameters executor.Parameters) bool {
		return false
	})
}

func (t *timer) OnRegisterExecutor(_ executor.Registry) error {
	return nil
}

func (t *timer) OnRegisterDataGen(reg dataGen.Registry) error {
	var generator = dataGen.NewGenerator(pluginName)
	generator.Add("", dataGen.New(
		[]string{},
		dataGen.Parameters{},
		dataGen.LangTypeCheck,
	))
	return reg.AddGenerator(pluginName, generator)
}

func (t *timer) OnRegisterLang(reg lang.Registry) error {
	return reg.Register(pluginName)
}
func NewTimer() plugins.Item {
	return &timer{}
}
