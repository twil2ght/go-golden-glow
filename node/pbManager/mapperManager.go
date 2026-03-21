package pbmanager

import (
	"fmt"
	"goldenglow/utils"
	"maps"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	sync "sync"

	"github.com/samber/lo"
	"google.golang.org/protobuf/proto"
)

type DMapper struct {
	fileDir string
	data    map[string]string
	mu      sync.RWMutex // 增加读写锁保证并发安全
}

type opt func(dm *DMapper)

// 移除全局实例，改为推荐创建实例的方式
var (
	BackExtraKeyWordsReversed = map[string]string{
		"[0xA]":      "A",
		"[0xB]":      "B",
		"[0xC]":      "C",
		"[0xD]":      "D",
		"[0xE]":      "E",
		"[0xF]":      "F",
		"[0xThing]":  "it",
		"[0xPerson]": "he",
	}
	defaultSourcePath string
	sourceFileName    = "mapper.pb"
	initMapper        = map[string]string{
		// "I":    "{[speaker]}",
		// "me":   "{[speaker]}",
		// "you":  "{[listener]}",
		// "my":   "{[speaker]} 's",
		// "your": "{[listener]} 's",
		"I":    "[speaker]",
		"me":   "[speaker]",
		"you":  "[listener]",
		"my":   "[speaker] 's",
		"your": "[listener] 's",
	}
	KeyWordMapper = map[string]bool{
		utils.Speaker:  true,
		utils.Listener: true,
	}
)

func init() {

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		panic(fmt.Errorf("node包初始化失败:无法获取当前文件路径"))
	}

	currentDir := filepath.Dir(currentFile)

	targetPath := filepath.Join(currentDir, sourceFileName)

	defaultSourcePath, _ = filepath.Abs(targetPath)
}

// NewDMapper 创建 DMapper 实例（替代全局 Dmapper）
func NewDMapper(options ...opt) (*DMapper, error) {
	dm := &DMapper{
		fileDir: defaultSourcePath,
		data:    make(map[string]string),
	}

	for _, option := range options {
		option(dm)
	}

	if err := dm.Init(); err != nil {
		return nil, err
	}

	return dm, nil
}

// Init 初始化 mapper，返回错误而非 panic
func (dm *DMapper) Init() error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	// 检查文件状态
	_, err := os.Stat(dm.fileDir)
	if err != nil {
		if os.IsNotExist(err) {
			// 文件不存在：初始化数据并写入
			maps.Copy(dm.data, initMapper)
			if err := dm.writeFile(); err != nil {
				return err // 返回错误，而非 panic
			}
			return nil // 初始化完成，无需后续读取
		}
		return err // 返回其他文件错误
	}

	// 文件存在：读取并反序列化
	data, err := os.ReadFile(dm.fileDir)
	if err != nil {
		return err
	}

	var mapper DynamicMapper
	if err := proto.Unmarshal(data, &mapper); err != nil {
		return err
	}

	dm.data = mapper.Data
	return nil
}
func (dm *DMapper) Show() {
	space := "============================================\n"
	fmt.Print(space)
	for k, v := range dm.data {
		fmt.Printf("[DynamicMapper] %s : %s\n", k, v)
	}
	fmt.Print(space)
}

// Set 设置键值对，返回错误
func (dm *DMapper) Set(k, v string) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	dm.data[k] = v
	return dm.writeFile()
}

// Get 获取所有数据的拷贝（只读锁优化）
func (dm *DMapper) Get() map[string]string {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	res := make(map[string]string, len(dm.data)) // 预分配容量，优化性能
	maps.Copy(res, dm.data)
	return res
}

// Del 删除键，返回错误
func (dm *DMapper) Del(k string) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	delete(dm.data, k)
	return dm.writeFile()
}

// writeFile 内部方法，需在加锁后调用
func (dm *DMapper) writeFile() error {
	// 检查文件路径是否为空
	if dm.fileDir == "" {
		return os.ErrInvalid
	}

	mapper := &DynamicMapper{
		Data: dm.data,
	}

	data, err := proto.Marshal(mapper)
	if err != nil {
		return err
	}

	// 写入文件，使用 0644 权限
	return os.WriteFile(dm.fileDir, data, 0644)
}
func (dm *DMapper) Encode(s string) string {
	change := true
	res := s
	for change {
		e := res
		parts := lo.Map(strings.Fields(res), func(e string, _ int) string {
			if v, ok := dm.data[e]; ok {
				return v
			}
			return e
		})
		res = strings.Join(parts, " ")
		change = e != res
	}
	return res
}
func (dm *DMapper) Decode(s string) string {
	change := true
	res := s
	for change {
		e := res
		parts := lo.Map(strings.Fields(res), func(e string, _ int) string {
			if v, ok := BackExtraKeyWordsReversed[e]; ok {
				return v
			}
			return e
		})
		res = strings.Join(parts, " ")
		change = e != res
	}
	return res
}

// WithSourceDir 自定义文件路径
func WithSourceDir(filePath string) opt {
	return func(dm *DMapper) {
		dm.fileDir = filePath
	}
}

var (
	dmapperInstance *DMapper
	dmapperOnce     sync.Once // 保证只初始化一次
	dmapperErr      error
)

// GetDmapper 获取全局 DMapper 实例，初始化失败会返回错误
func GetDmapper() (*DMapper, error) {
	dmapperOnce.Do(func() {
		dmapperInstance, dmapperErr = NewDMapper()
	})
	return dmapperInstance, dmapperErr
}
