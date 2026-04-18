package datagen

import (
	"fmt"
	"goldenglow/pkg/node/handler"
	"goldenglow/pkg/registry"
	"goldenglow/utils"
	"path/filepath"
	"sort"
)

type DataExtraType int
type DataExtraPosition int

const (
	AsExecutor DataExtraType = iota
	AsChecker
	AsExtractor
)
const (
	AsT DataExtraPosition = iota
	AsR
)

var (
	RootDir                   = filepath.Join(utils.RootDir, "data")
	jsonExt                   = ".json"
	ReflectDataExtraTypeToKey = map[DataExtraType]string{
		AsExecutor:  handler.NodeExecutor,
		AsChecker:   handler.NodeChecker,
		AsExtractor: handler.NodeExtractor,
	}
	ReflectDataExtraTypeToPosition = map[DataExtraType]DataExtraPosition{
		AsExecutor:  AsR,
		AsChecker:   AsT,
		AsExtractor: AsT,
	}
	KeyNamespace = "namespace"
)

type Data interface {
	T() []string
	R() []string
	BuildExtra(namespace string)
}
type Provider interface {
	Add(name string, item Data)
	Run(namespace string)
}
type Generator interface {
	AddProvider(namespace string, item Provider)
	Run()
}
type data struct {
	t, r              []string
	paramToBuildExtra map[string]string
	extraType         DataExtraType
}

func (d *data) T() []string {
	return d.t
}
func (d *data) R() []string {
	return d.r
}
func (d *data) BuildExtra(namespace string) {
	kind := ReflectDataExtraTypeToKey[d.extraType]
	res := fmt.Sprintf("%s [%s:%s]", kind, KeyNamespace, namespace)
	params := d.paramToBuildExtra
	position := ReflectDataExtraTypeToPosition[d.extraType]
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, k := range keys {
		res = fmt.Sprintf("%s [%s:%s]", res, k, params[k])
	}
	switch position {
	case AsT:
		d.t = append(d.t, res)
	case AsR:
		d.r = append(d.r, res)
	}
}

// NewData is the Data constructor
func NewData(t, r []string, params map[string]string, extraType DataExtraType) Data {
	return &data{
		t:                 t,
		r:                 r,
		paramToBuildExtra: params,
		extraType:         extraType,
	}
}

type provider struct {
	items registry.Interface[Data]
}

func (p *provider) Run(namespace string) {
	p.items.Range(func(name string, item Data) bool {
		item.BuildExtra(namespace)
		_ = saveToJsonWithNamespace(item, makePath(namespace, name), namespace)
		return true
	})
}

func (p *provider) Add(name string, item Data) {
	p.items.Register(name, item)
}

// NewProvider is the Provider constructor
func NewProvider() Provider {
	return &provider{
		items: registry.New[Data](),
	}
}

type generator struct {
	items registry.Interface[Provider]
}

func (g *generator) Run() {
	g.items.Range(func(namespace string, p Provider) bool {
		p.Run(namespace)
		return true
	})
}

func (g *generator) AddProvider(namespace string, item Provider) {
	g.items.Register(namespace, item)
}

// NewGenerator is the Generator constructor
func NewGenerator() Generator {
	return &generator{
		items: registry.New[Provider](),
	}
}
