package timer

import (
	"fmt"
	"goldenglow/m"
	"goldenglow/node"
	"goldenglow/plugin"
	"goldenglow/plugins/checker"
	"sync"
	"time"
)

// TODO 整个都需要重新查看
// ======================== 常量定义 ========================
// 事件类型
const (
	EventNow    = "[now]"    // 获取当前时间事件
	EventOnce   = "[once]"   // 单次定时任务事件
	EventTicker = "[ticker]" // 周期定时任务事件
	EventStop   = "[stop]"   // 停止定时任务事件
)

// 参数常量
var (
	ParamDelay   = "[delay]"    // 延迟时间(秒)
	ParamPeriod  = "[period]"   // 周期时间(秒)
	ParamTaskID  = "[taskID]"   // 任务ID
	ParamEvent   = plugin.Event // 事件类型
	ParamUserID  = "[userID]"   // 用户ID
	ParamContent = "[content]"  // 任务内容
)

// ======================== 核心结构体 ========================
// Engine 定时器引擎（管理所有定时任务）
type Engine struct {
	tasks map[string]*Task // 任务ID -> 任务
	mu    sync.RWMutex     // 并发安全锁
}

// Task 定时任务结构体
type Task struct {
	timer    *time.Timer   // 单次定时器
	ticker   *time.Ticker  // 周期定时器
	stopChan chan struct{} // 停止信号
	closed   bool          // 是否已关闭
}

// Context 插件上下文
type Context struct {
	UserID  string
	Delay   int
	Period  int
	TaskID  string
	Content string
	Event   string
	params  m.H
}

// TimerNode 插件节点
type TimerNode struct{ node.Base }

// ======================== 全局单例 ========================
var engine = &Engine{
	tasks: make(map[string]*Task),
}

// ======================== TimerEngine 核心方法 ========================
func (e *Engine) NowTime(c *Context) string {
	now := time.Now()
	// 格式化返回：秒 分 时 日 周 月 年
	return fmt.Sprintf(
		"%d %d %d %d %d %d %d",
		now.Second(),
		now.Minute(),
		now.Hour(),
		now.Day(),
		now.Weekday(),
		now.Month(),
		now.Year(),
	)
}

func (e *Engine) OnceTask(c *Context) {
	if c.Delay <= 0 || c.TaskID == "" {
		fmt.Printf("[Timer] Invalid once task: delay=%d, taskID=%s\n", c.Delay, c.TaskID)
		return
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	// 停止已存在的同名任务
	if oldTask, ok := e.tasks[c.TaskID]; ok {
		oldTask.timer.Stop()
		close(oldTask.stopChan)
	}

	// 创建新任务
	stopChan := make(chan struct{})
	timer := time.AfterFunc(time.Duration(c.Delay)*time.Second, func() {
		select {
		case <-stopChan:
			return
		default:
			fmt.Printf("[Timer] Once task [%s] executed: %s\n", c.TaskID, c.Content)
			// 任务执行完成后清理
			e.mu.Lock()
			delete(e.tasks, c.TaskID)
			e.mu.Unlock()
		}
	})

	// 存储任务
	e.tasks[c.TaskID] = &Task{
		timer:    timer,
		stopChan: stopChan,
	}

	fmt.Printf("[Timer] Once task [%s] created, delay %ds\n", c.TaskID, c.Delay)
}

// TickerTask 执行周期定时任务
func (e *Engine) TickerTask(c *Context) {
	if c.Period <= 0 || c.TaskID == "" {
		fmt.Printf("[Timer] Invalid ticker task: period=%d, taskID=%s\n", c.Period, c.TaskID)
		return
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	// 停止已存在的同名任务
	if oldTask, ok := e.tasks[c.TaskID]; ok {
		oldTask.ticker.Stop()
		close(oldTask.stopChan)
	}

	// 创建新任务
	stopChan := make(chan struct{})
	ticker := time.NewTicker(time.Duration(c.Period) * time.Second)

	// 启动协程执行周期任务
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				fmt.Printf("[Timer] Ticker task [%s] executed: %s\n", c.TaskID, c.Content)
			case <-stopChan:
				fmt.Printf("[Timer] Ticker task [%s] stopped\n", c.TaskID)
				return
			}
		}
	}()

	// 存储任务
	e.tasks[c.TaskID] = &Task{
		ticker:   ticker,
		stopChan: stopChan,
	}

	fmt.Printf("[Timer] Ticker task [%s] created, period %ds\n", c.TaskID, c.Period)
}

// StopTask 停止指定任务
func (e *Engine) StopTask(c *Context) {
	if c.TaskID == "" {
		return
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	task, ok := e.tasks[c.TaskID]
	if !ok {
		fmt.Printf("[Timer] Task [%s] not found\n", c.TaskID)
		return
	}

	if !task.closed {
		if task.timer != nil {
			task.timer.Stop()
		}
		if task.ticker != nil {
			task.ticker.Stop()
		}
		close(task.stopChan)
		task.closed = true
	}

	delete(e.tasks, c.TaskID)
	fmt.Printf("[Timer] Task [%s] stopped and removed\n", c.TaskID)
}

// ======================== HandleCtx 接口实现 ========================
func (c *Context) Get() m.H {
	if c.params == nil {
		c.params = make(m.H)
	}
	c.params[ParamUserID] = c.UserID
	c.params[ParamDelay] = c.Delay
	c.params[ParamPeriod] = c.Period
	c.params[ParamTaskID] = c.TaskID
	c.params[ParamContent] = c.Content
	c.params[ParamEvent] = c.Event
	return c.params
}

func (c *Context) Set(params m.H) {
	if params == nil {
		return
	}
	c.params = params

	if userID, ok := params[ParamUserID].(string); ok {
		c.UserID = userID
	}
	if delay, ok := params[ParamDelay].(int); ok {
		c.Delay = delay
	}
	if period, ok := params[ParamPeriod].(int); ok {
		c.Period = period
	}
	if taskID, ok := params[ParamTaskID].(string); ok {
		c.TaskID = taskID
	}
	if content, ok := params[ParamContent].(string); ok {
		c.Content = content
	}
	if event, ok := params[ParamEvent].(string); ok {
		c.Event = event
	}
}

// ======================== 事件处理器 ========================
var Handlers = map[string]plugin.Handler{
	EventNow: func(ctx plugin.HandlerCtx) {
		c, ok := ctx.(*Context)
		if !ok {
			return
		}
		timeStr := engine.NowTime(c)
		fmt.Printf("[Timer] Current time: %s\n", timeStr)
	},
	EventOnce: func(ctx plugin.HandlerCtx) {
		c, ok := ctx.(*Context)
		if !ok {
			return
		}
		engine.OnceTask(c)
	},
	EventTicker: func(ctx plugin.HandlerCtx) {
		c, ok := ctx.(*Context)
		if !ok {
			return
		}
		engine.TickerTask(c)
	},
	EventStop: func(ctx plugin.HandlerCtx) {
		c, ok := ctx.(*Context)
		if !ok {
			return
		}
		engine.StopTask(c)
	},
}

// ======================== 参数配置 ========================
// 格式: [Timer] userID & delay & period & taskID & content & event
var ParamKeys = map[string]int{
	ParamUserID:  0,
	ParamDelay:   1,
	ParamPeriod:  2,
	ParamTaskID:  3,
	ParamContent: 4,
	ParamEvent:   5,
}
var (
	head = "[timer]"
)
var ParamConfig = &plugin.ParamCfg{
	Head:        head,
	ParamLength: 6,
}

// ======================== 翻译API ========================
var LangAPI = plugin.TRGroup{
	"get_time": {
		TV: plugin.NV{"[say] [0x01] get current time"},
		RV: plugin.NV{"[Timer] [0x01] & 0 & 0 & time_now & get_time & [now]"},
	},
	"once_task": {
		TV: plugin.NV{"[say] [0x01] create once task after [0x02]s: [0x03]"},
		RV: plugin.NV{"[Timer] [0x01] & [0x02] & 0 & [0x04] & [0x03] & [once]"},
	},
	"ticker_task": {
		TV: plugin.NV{"[say] [0x01] create ticker task every [0x02]s: [0x03]"},
		RV: plugin.NV{"[Timer] [0x01] & 0 & [0x02] & [0x04] & [0x03] & [ticker]"},
	},
	"stop_task": {
		TV: plugin.NV{"[say] [0x01] stop task [0x02]"},
		RV: plugin.NV{"[Timer] [0x01] & 0 & 0 & [0x02] & stop & [stop]"},
	},
	//for checker plugin
	WithinRangeEvent: {
		TV: plugin.NV{
			"check if [0x02] < [0x01] < [0x03]",
			fmt.Sprintf("%s %s [0x01] & [0x02] & [0x03]", checker.Head, WithinRangeEvent),
		},
		RV: plugin.NV{
			"[0x02] < [0x01] < [0x03]",
		},
	},
}

// ======================== Node 实现 ========================
func (t *TimerNode) Execute() error {
	return plugin.Execute(
		head,
		t.Parse,
		&plugin.HandlerCtxBase{},
		Handlers,
	)
}

func (t *TimerNode) Parse() (m.H, error) {
	return plugin.ParseFunc(t.Base, ParamConfig, ParamKeys)
}

func nodeCreator(base node.Base) node.Item {
	return &TimerNode{Base: base}
}

func New(e *Engine) plugin.Item {
	return &plugin.Base{
		ParamCfg:    ParamConfig,
		NodeCreator: nodeCreator,
		LangAPI:     LangAPI,
		Name:        "Timer",
	}
}
func NewEngine() *Engine {
	return &Engine{
		tasks: make(map[string]*Task),
	}
}

// ShutdownTimer 插件销毁，清理所有任务
func ShutdownTimer() {
	engine.mu.Lock()
	defer engine.mu.Unlock()

	for _, task := range engine.tasks {
		if task.timer != nil {
			task.timer.Stop()
		}
		if task.ticker != nil {
			task.ticker.Stop()
		}
		if !task.closed {
			close(task.stopChan)
		}
	}
	engine.tasks = nil
	fmt.Println("[Timer] Plugin shutdown, all tasks stopped")
}
