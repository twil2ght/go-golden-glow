package dictionary

import (
	"encoding/json"
	"fmt"
	"goldenglow/m"
	"goldenglow/node"
	"goldenglow/plugin"
	"os"
	"path/filepath"
	"strings"
)

type Engine struct {
	Table    map[string]struct{}
	dataPath string
}

var engine *Engine

var (
	FilePath = "./data/table.json"
	head     = "[Dictionary]"
	EventSet = "[HSet]"
	Word     = "word"
	offsets  = []string{
		"is",
		"[GG] 's",
		"[GG]",
		"and",
		"object",
		"adjective",
		"noun",
		"verb",
		"action",
	}
)

// Set 添加新单词到字典
func (d *Engine) Set(newWord string) {
	if newWord != "" && len(strings.Fields(newWord)) == 1 { // 空单词不处理
		d.Table[newWord] = struct{}{}
	}
}

// LoadArchive 从 table.json 加载字典数据
func (d *Engine) LoadArchive() {
	// 确保存储目录存在
	dir := filepath.Dir(FilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Printf("[Dictionary] 创建目录失败: %v\n", err)
		return
	}

	// 读取文件
	data, err := os.ReadFile(FilePath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("[Dictionary] 字典文件不存在，创建新文件: %s\n", FilePath)
			// 首次运行创建空文件
			if err := d.Save(); err != nil {
				fmt.Printf("[Dictionary] 创建空字典文件失败: %v\n", err)
			}
			return
		}
		fmt.Printf("[Dictionary] 读取字典文件失败: %v\n", err)
		return
	}

	// JSON 解析
	var words []string
	if err := json.Unmarshal(data, &words); err != nil {
		fmt.Printf("[Dictionary] 解析字典文件失败: %v\n", err)
		return
	}

	// 加载到 Table 中
	for _, word := range words {
		d.Set(word)
	}
	fmt.Printf("[Dictionary] 成功加载 %d 个单词\n", len(words))
}

// Save 将字典数据保存到 table.json
func (d *Engine) Save() error {
	// 将 map 转换为字符串切片（JSON 序列化更友好）
	words := make([]string, 0, len(d.Table))
	for word := range d.Table {
		words = append(words, word)
	}

	// JSON 序列化
	data, err := json.MarshalIndent(words, "", "  ") // 格式化输出，便于阅读
	if err != nil {
		return fmt.Errorf("序列化字典失败: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(FilePath, data, 0644); err != nil {
		return fmt.Errorf("写入字典文件失败: %w", err)
	}

	fmt.Printf("[Dictionary] 成功保存 %d 个单词到 %s\n", len(words), FilePath)
	return nil
}

type DictionaryNode struct {
	node.Base
}

var (
	paramKeys = map[string]int{
		Word:         0,
		plugin.Event: 1,
	}
	paramConfig = &plugin.ParamCfg{
		Head:        head,
		ParamLength: 3,
	}
	handlers = map[string]plugin.Handler{
		EventSet: func(ctx plugin.HandlerCtx) {

		},
	}
)

func (dn *DictionaryNode) Execute() error {
	return plugin.Execute(head, dn.Parse, &plugin.HandlerCtxBase{}, handlers)
}

func (dn *DictionaryNode) Parse() (m.H, error) {
	return plugin.ParseFunc(dn.Base, paramConfig, paramKeys)
}

var langAPI = map[string]plugin.TR{
	"HSet": {
		TV: plugin.NV{"[0x01] is [0x02]"},
		RV: plugin.NV{fmt.Sprintf("%s [0x01] & %s", head, EventSet)},
	},
}

func nodeCreator(base node.Base) node.Item {
	return &DictionaryNode{
		Base: base,
	}
}
func New(e *Engine) plugin.Item {
	if e != nil {
		engine = e
	}
	engine.LoadArchive()
	return &plugin.Base{
		ParamCfg:    paramConfig,
		NodeCreator: nodeCreator,
		LangAPI:     langAPI,
		Name:        head,
	}
}
func NewEngine(dataPath string) *Engine {
	return &Engine{
		Table:    make(map[string]struct{}),
		dataPath: dataPath,
	}
}
func ShutDown() {
	engine.Save()
}
