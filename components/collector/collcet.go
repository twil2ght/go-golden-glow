package collector

import (
	"goldenglow/components"
	"goldenglow/components/receiver"
	"goldenglow/config"
	"goldenglow/m"
	"goldenglow/pkg/log"
	"goldenglow/storage"
	"goldenglow/utils"
	"os"
	"time"
)

var path = utils.RootDir + "/archive/history"
var logger = log.Default()

type collector struct {
	rcv          chan components.Message
	userTag      string
	selfTag      string
	msgCollected []string
}

func (c *collector) Save() error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return err
	}
	fileName := time.Now().Format("2006-01-02 15-04-05")
	filePath := path + "/" + fileName + ".json"
	logger.Debug("saving history to file", "file", filePath)
	return storage.SaveAsJson(filePath, c.msgCollected)
}

func (c *collector) OnRegisterReceiver(reg receiver.Registry) error {
	var subs = m.Hash{
		c.userTag: struct{}{},
		c.selfTag: struct{}{},
	}
	//for sub := range reg.Subscriptions() {
	//	logger.Info("subscription", "tag", sub)
	//}
	return reg.Register("collector", subs, func() chan<- components.Message {
		return c.rcv
	})
}
func (c *collector) SetSource(userTag string, selfTag string) {
	c.userTag = userTag
	c.selfTag = selfTag
}
func (c *collector) Run() {
	for msg := range c.rcv {
		var realTag string
		switch msg.Tag() {
		case c.userTag:
			realTag = config.User
		case c.selfTag:
			realTag = config.GG
		}
		value := realTag + ":" + msg.Value()
		c.msgCollected = append(c.msgCollected, value)
	}
}

func New() Instance {
	return &collector{
		rcv: make(chan components.Message, 10),
	}
}
