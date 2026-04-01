package source

import (
	"goldenglow/components"
	"goldenglow/m"
	"goldenglow/pkg/log"
	"goldenglow/utils"
)

var (
	logger = log.Default()
)

type registry struct {
	sources    map[string]Source
	mainstream chan components.Message
}

func (r *registry) Tags() m.Hash {
	tags := make(m.Hash)
	for tag := range r.sources {
		tags[tag] = struct{}{}
	}
	return tags
}

func (r *registry) C() <-chan components.Message {
	//logger.Info("InputSource:start mainstream", "source_amount", len(r.sources))
	for tag, ch := range r.sources {
		go func(source Source, tag string) {
			msgCount := 0
			for msg := range source.C() {
				msgCount++
				logger.Info("InputSource:received from source", "tag", tag, "msgNum", msgCount, "msg", msg)
				r.mainstream <- components.NewMsg(msg, tag)
				logger.Info("InputSource:sent to mainstream", "tag", tag, "msgNum", msgCount)
			}
			logger.Info("InputSource:source channel closed", "tag", tag, "totalMessages", msgCount)
		}(ch, tag)
	}

	return r.mainstream
}

func (r *registry) Register(pluginName, tag string, source Source) error {
	if err := utils.NotNull(
		"plugin name", pluginName,
		"tag", tag,
		"source", source,
	); err != nil {
		return err
	}
	key := pluginName + ":" + tag
	r.sources[key] = source
	logger.Info("InputSource:register source", "pluginName", pluginName, "tag", tag)
	return nil
}
func NewRegistry() Registry {
	return &registry{
		sources:    make(map[string]Source),
		mainstream: make(chan components.Message, 1000),
	}
}
