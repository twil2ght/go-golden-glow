package timer

import (
	"fmt"
	"goldenglow/m"
	"goldenglow/pkg/datagen"
	"goldenglow/pkg/messageQueue"
	"goldenglow/pkg/node/handler"
	"goldenglow/pkg/registry"
	"goldenglow/pkg/variable"
	"goldenglow/plugin"
	"time"
)

func init() {
	plugin.DefaultManager.Register(name, NewTimer())
}

const (
	name = "timer"

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

	keyDuration = "duration"
	keyItem     = "item"
)

type timer struct {
	tickerManager registry.Interface[*time.Timer]
	tickerChannel chan string
	shutdown      chan struct{}
}

func (t *timer) OnRegisterMsgProvider(reg messageQueue.Manager) {
	reg.Add(name, t.tickerChannel)
}
func (t *timer) tickerProcess(item string, duration int) {
	d := time.Duration(duration) * time.Second

	oldTimer, err := t.tickerManager.Get(item)
	if err == nil {
		oldTimer.Stop()
		fmt.Printf("[timer] timer stop old timer:%s\n", item)
	}

	newTimer := time.AfterFunc(d, func() {
		t.TickerCallback(item)
	})
	t.tickerManager.Unregister(item)
	t.tickerManager.Register(item, newTimer)
}
func (t *timer) TickerCallback(item string) {
	t.tickerManager.Unregister(item)
	// implement other logic below
	t.tickerChannel <- fmt.Sprintf("[ticker:timeout] %s", item)
}
func (t *timer) OnRegisterExecutor(reg handler.Executor[handler.ExecuteHandler]) {
	reg.Handlers().Register(name, func(parameters handler.Parameters) {
		var (
			item, _        = parameters.Get(keyItem)
			rawDuration, _ = parameters.Get(keyDuration)
			duration       int
		)
		_, err := fmt.Sscanf(rawDuration, "%d", &duration)
		if err != nil {
			return
		}
		t.tickerProcess(item, duration)
	})
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

func (t *timer) OnRegisterExtractor(reg handler.Executor[handler.ExtractorHandler]) {
	reg.Handlers().Register(name, func(parameters handler.Parameters) variable.ValueMap {
		mode, err := parameters.Get(keyMode)
		if err != nil {
			return nil
		}
		typeVal, err := parameters.Get(keyType)
		if err != nil {
			return nil
		}

		value, err := t.process(mode, typeVal)
		if err != nil {
			return nil
		}
		return variable.NewValueMap(m.Hash{value: struct{}{}})
	})
}

func (t *timer) OnRegisterDataGen(gen datagen.Generator) {
	provider := datagen.NewProvider()
	provider.Add("ticker", datagen.NewData(
		[]string{"[ticker] $1 -> $3 @Caller $2"},
		[]string{},
		map[string]string{
			keyDuration: "$3",
			keyItem:     "$1",
		},
		datagen.AsExecutor,
	))
	provider.Add("fetch_time", datagen.NewData(
		[]string{"[time] get time"},
		[]string{"[time] time -> $1"},
		map[string]string{
			keyMode: modeFetch,
			keyType: typeTime,
			KeyDist: "$1",
		},
		datagen.AsExtractor,
	))
	provider.Add("fetch_hour", datagen.NewData(
		[]string{"[time] get hour"},
		[]string{"[time] hour -> $1"},
		map[string]string{
			keyMode: modeFetch,
			keyType: typeHour,
			KeyDist: "$1",
		},
		datagen.AsExtractor,
	))
	provider.Add("fetch_minute", datagen.NewData(
		[]string{"[time] get minute"},
		[]string{"[time] minute -> $1"},
		map[string]string{
			keyMode: modeFetch,
			keyType: typeMinute,
			KeyDist: "$1",
		},
		datagen.AsExtractor,
	))
	provider.Add("fetch_date", datagen.NewData(
		[]string{"[time] get date"},
		[]string{"[time] date -> $1"},
		map[string]string{
			keyMode: modeFetch,
			keyType: typeDate,
			KeyDist: "$1",
		},
		datagen.AsExtractor,
	))
	gen.AddProvider(name, provider)
}

func (t *timer) Init() {}

func (t *timer) Shutdown() {}

func NewTimer() plugin.Interface {
	return &timer{
		tickerManager: registry.New[*time.Timer](),
		tickerChannel: make(chan string, 1000),
		shutdown:      make(chan struct{}),
	}
}
