package MarkUnknown

import (
	"fmt"
	"goldenglow/m"
	"goldenglow/pkg/database"
	"goldenglow/pkg/datagen"
	"goldenglow/pkg/node/handler"
	"goldenglow/pkg/variable"
	"goldenglow/plugin"
	"strings"
)

func init() {
	plugin.DefaultManager.Register(Name, &UnknownMarker{repo: database.DefaultRedisRepo()})
}

type UnknownMarker struct {
	repo database.RedisRepository
}

func (p *UnknownMarker) Init() {

}

func (p *UnknownMarker) Shutdown() {

}

const (
	TypeMsg = "msg"
	Name    = "unknown_marker"
	keyDist = "dist"
)

func (p *UnknownMarker) OnRegisterDataGen(gen datagen.Generator) {
	provider := datagen.NewProvider()
	provider.Add("locate unknown", datagen.NewData(
		[]string{"[mark] $1 @Caller $2"},
		[]string{"[mark] $2 @ $1 -> $3"},
		m.Map[string]{
			TypeMsg: "$1",
			keyDist: "$3",
		},
		datagen.AsExtractor,
	),
	)
	provider.Add("legal one", datagen.NewData(
		[]string{"[mark] $1 @Caller $2"},
		[]string{"[mark] $2 @ $1 ->"},
		m.Map[string]{
			TypeMsg: "$1",
		},
		datagen.AsChecker,
	),
	)
	gen.AddProvider(Name, provider)
}
func (p *UnknownMarker) OnRegisterChecker(reg handler.Executor[handler.CheckHandler]) {
	reg.Handlers().Register(Name, func(parameters handler.Parameters) bool {
		var (
			msg, _ = parameters.Get(TypeMsg)
		)
		for v := range strings.SplitSeq(msg, " ") {
			_, err := p.repo.HGet("[Units] " + v)
			if err != nil {
				return false
			}
		}
		return true
	})
}
func (p *UnknownMarker) OnRegisterExtractor(reg handler.Executor[handler.ExtractorHandler]) {
	reg.Handlers().Register(Name, func(parameters handler.Parameters) variable.ValueMap {
		var (
			msg, _ = parameters.Get(TypeMsg)
		)
		for v := range strings.SplitSeq(msg, " ") {
			_, err := p.repo.HGet("[Units] " + v)
			if err != nil {
				fmt.Printf("could not find unit %s\n", "[Units] "+v)
				return variable.NewValueMap(m.Hash{v: struct{}{}})
			}
		}
		return nil
	})
}
