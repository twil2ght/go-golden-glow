package receiver

import (
	"errors"
	"goldenglow/components"
	"goldenglow/components/source"
	"goldenglow/m"
	"goldenglow/pkg/log"
	"goldenglow/utils"
	"sync"
)

var logger = log.Default()

// registry 消息订阅路由注册表
type registry struct {
	chs        map[string]chan<- components.Message // tag -> 接收者通道
	route      map[string]m.Hash                    // 订阅标签(tag) -> 接收者tag集合（路由表）
	mainstream source.MainStream                    // 全局消息源
	mu         sync.RWMutex
}

func (r *registry) Subscriptions() m.Hash {
	var subs = make(m.Hash)
	for sub := range r.route {
		subs[sub] = struct{}{}
	}
	return subs
}

func NewRegistry(ms source.MainStream, subs m.Hash) Registry {
	var routes = make(map[string]m.Hash)
	for sub := range subs {
		routes[sub] = make(m.Hash)
	}
	return &registry{
		chs:        make(map[string]chan<- components.Message),
		mainstream: ms,
		route:      routes,
	}
}

// Register 注册订阅者：tag=接收者唯一标识，subscribeTo=要订阅的消息标签，r=接收器函数
func (r *registry) Register(tag string, subscribeTo m.Hash, rcv Receiver) error {
	// 参数非空校验
	if err := utils.NotNull(
		"tag", tag,
		"subscribeTo", subscribeTo,
		"receiver", rcv,
	); err != nil {
		return err
	}

	// 加写锁：并发安全写入map
	r.mu.Lock()
	defer r.mu.Unlock()

	// 1. 存储接收者通道
	ch := rcv()
	if ch == nil {
		return errors.New("receiver channel is nil")
	}
	r.chs[tag] = ch
	// 2. 遍历要订阅的所有消息标签，构建路由表
	for subTag := range subscribeTo {
		if r.route[subTag] == nil {
			r.route[subTag] = make(m.Hash)
		}
		r.route[subTag][tag] = struct{}{}
	}

	return nil
}

// Start 启动消息监听协程：从全局消息流消费消息，按路由推送给订阅者
func (r *registry) Start() {
	go func() {
		// 持续消费全局消息流
		for msg := range r.mainstream.C() {
			msgTag := msg.Tag() // 消息自带的标签
			logger.Debug("consuming message", "tag", msgTag, "message", msg.Value())
			// 加读锁：并发安全读取路由表
			r.mu.RLock()
			// 根据消息标签，找到所有订阅该标签的接收者tag
			receivers, exists := r.route[msgTag]
			if !exists {
				r.mu.RUnlock()
				logger.Debug("no subscribers for message tag", "tag", msgTag)
				continue
			}
			logger.Debug("subscribers for message tag", "tag", msgTag, "subscribers_amount", len(receivers))
			// 遍历所有订阅者，推送消息
			for receiverTag := range receivers {
				ch, ok := r.chs[receiverTag]
				if !ok {
					logger.Debug("receiver channel not found", "tag", receiverTag)
					continue
				}

				// 解锁：发送消息是阻塞操作，必须提前释放锁
				r.mu.RUnlock()

				// 非阻塞发送（避免通道满导致协程卡住）
				select {
				case ch <- msg:
					logger.Debug("sent message to receiver", "tag", receiverTag, "message", msg.Value())
				default:
					// 通道已满，丢弃消息（可替换为日志）
				}

				// 重新加锁，继续遍历
				r.mu.RLock()
			}
			r.mu.RUnlock()
		}
	}()
}
