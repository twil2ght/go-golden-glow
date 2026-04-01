package timer

import (
	"fmt"
	"goldenglow/dataGen"
	"goldenglow/executor"
	"goldenglow/executor/extractor"
	"goldenglow/lang"
	"goldenglow/pkg/log"
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

var logger = log.Default()

const (
	pluginName = "timer"

	//key
	keyMode = "mode"
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
	now := time.Now()
	switch arg {
	case typeTime:
		return fmt.Sprintf("%02d:%02d:%02d", now.Hour(), now.Minute(), now.Second()), nil
	case typeDate:
		return fmt.Sprintf("%02d:%02d", now.Month(), now.Day()), nil
	case typeHour:
		return fmt.Sprintf("%02d", now.Hour()), nil
	case typeMinute:
		return fmt.Sprintf("%02d", now.Minute()), nil
	case typeSecond:
		return fmt.Sprintf("%02d", now.Second()), nil
	case typeYear:
		return fmt.Sprintf("%02d", now.Year()), nil
	case typeMonth:
		return fmt.Sprintf("%02d", now.Month()), nil
	case typeDay:
		return fmt.Sprintf("%02d", now.Day()), nil
	default:
		return "", fmt.Errorf("unknown timer type: %s", arg)
	}
}

func (t *timer) OnRegisterExtractor(reg extractor.Registry) error {
	return reg.Register(pluginName, func(params executor.Parameters) (variable.Item, error) {
		if err := executor.Validate(params, keyMode, keyType, KeyDist); err != nil {
			return nil, err
		}

		value, err := t.process(params[keyMode], params[keyType])
		if err != nil {
			logger.Error("error processing timer", "error", err)
			return nil, err
		}
		logger.Debug("timer processing value", "value", value)
		return variable.New(params[KeyDist], value), err
	})
}

func (t *timer) Name() string {
	return pluginName
}

func (t *timer) OnRegisterExecutor(_ executor.Registry) error {
	return nil
}

func (t *timer) OnRegisterDataGen(reg dataGen.Registry) error {
	var generator = dataGen.NewGenerator(pluginName)
	generator.Add("fetch_time", dataGen.SNew(
		"check what is the time now",
		"the time is $1 now",
		dataGen.Parameters{
			keyMode: modeFetch,
			keyType: typeTime,
			KeyDist: "$1",
		},
		dataGen.LangTypeExtractor,
	))
	generator.Add("fetch_hour", dataGen.SNew(
		"check what is the time now",
		"the hour is $1 o'clock now",
		dataGen.Parameters{
			keyMode: modeFetch,
			keyType: typeHour,
			KeyDist: "$1",
		},
		dataGen.LangTypeExtractor,
	))
	return reg.AddGenerator(pluginName, generator)
}

func (t *timer) OnRegisterLang(reg lang.Registry) error {
	return reg.Register(pluginName)
}
func NewTimer() plugins.Item {
	return &timer{}
}
