package betterspeak

import (
	"fmt"
	"goldenglow/m"
	"goldenglow/node"
	"goldenglow/plugin"
	"goldenglow/plugins"
	"goldenglow/storage"
	"strings"
	"sync"
	"time"
)

type Speaker interface {
	Send(msg string)
	Register(ch chan string)
	Receive() string
}
type speakerBase struct {
	channel chan string
	db      storage.KVLite
}

func (s *speakerBase) Receive() string {
	return <-s.channel
}

func (s *speakerBase) Register(ch chan string) {
	s.channel = ch
}

func (s *speakerBase) Send(str string) {
	s.channel <- s.decode(str)
}
func (s *speakerBase) decode(str string) string {
	res := make([]string, 0)
	for _, t := range strings.Fields(str) {
		if key, err := s.db.Keyof(t); err == nil {
			res = append(res, key)
		} else {
			res = append(res, t)
		}
	}
	return strings.Join(res, " ")
}
func NewEngine(db storage.KVLite) (Speaker, error) {
	if db == nil {
		return nil, fmt.Errorf("speaker init: db is nil")
	}
	return &speakerBase{
		channel: make(chan string),
		db:      db,
	}, nil
}

type SpeakerNode struct {
	node.Base
}

var (
	Head       = "[BetterSpeaker]"
	EventSpeak = "[Send]"

	Msg       = "msg"
	UserID    = "userID"
	ParamKeys = map[string]int{
		Msg:          0,
		plugin.Event: 1,
		UserID:       2,
	}
	Handlers = map[string]plugin.Handler{
		EventSpeak: func(ctx plugin.HandlerCtx) {
			msg, ok := ctx.Get()[Msg].(string)
			if !ok {
				return
			}
			engine.Send(msg)
		},
	}
)
var engine Speaker

func (s *SpeakerNode) Execute() error {
	return plugin.Execute(Head, s.Parse, &plugin.HandlerCtxBase{}, Handlers)
}

func (s *SpeakerNode) Parse() (m.H, error) {
	return plugin.ParseFunc(s.Base, ParamConfig, ParamKeys)
}

var ParamConfig = &plugin.ParamCfg{
	Head:        Head,
	ParamLength: 3,
}

var LangAPI = map[string]plugin.TR{
	"speak": {
		TV: plugin.NV{"[GG] should say [0x01] to [0x02]"},
		RV: plugin.NV{"[BetterSpeaker] [0x01] & [Send] & [0x02]"},
	},
}

func nodeCreator(base node.Base) node.Item {
	return &SpeakerNode{
		Base: base,
	}
}

func New(speaker Speaker) plugin.Item {
	if speaker != nil {
		engine = speaker
	}
	return &plugin.Base{
		ParamCfg:    ParamConfig,
		NodeCreator: nodeCreator,
		LangAPI:     LangAPI,
		Name:        Head,
	}
}

// Decorate extra
func Decorate(str string, userId int) string {
	return fmt.Sprintf("[say] [GG] to %d : %s //", userId, str)
}

type Input struct {
	Val          string
	Dec          pinkcat.Decorator
	NeedResponse bool
}
type Scheduler struct {
	Queue        []string
	InputQueue   []*Input
	Runner       plugins.Runner
	userid       int
	ControllerMu *sync.Mutex
}

func NewScheduler(runner plugins.Runner, userid int) *Scheduler {
	return &Scheduler{
		Queue:        make([]string, 0),
		InputQueue:   make([]*Input, 0),
		Runner:       runner,
		userid:       userid,
		ControllerMu: &sync.Mutex{},
	}
}
func (s *Scheduler) Set(str string) {
	if str != "" {
		s.InputQueue = append(s.InputQueue, &Input{
			Val: str,
		})
	}
}
func (s *Scheduler) Exhaust() *Input {
	if len(s.InputQueue) == 0 {
		return nil
	}
	res := s.InputQueue[0]
	s.InputQueue = s.InputQueue[1:]
	return res
}
func (s *Scheduler) Collector(speaker Speaker) {
	checkInterval := 1 * time.Second
	go func() {
		ticker := time.NewTicker(checkInterval)
		defer ticker.Stop()
		for {
			str := speaker.Receive()
			if str != "" {
				s.Queue = append(s.Queue, str)
				s.InputQueue = append(s.InputQueue, &Input{
					Val: str,
					Dec: func(str string) string {
						return Decorate(str, s.userid)
					},
				})
			}
		}
	}()
}
