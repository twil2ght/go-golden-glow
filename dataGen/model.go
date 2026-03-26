package dataGen

type Item interface {
	Triggers() []string
	Results() []string
	LangType() LangType
	Params() Parameters
}
type Generator interface {
	Add(name string, langItem Item)
	Run() error
}
type Registry interface {
	RunAll() error
	AddGenerator(pluginName string, generator Generator) error
}
type LangType string
type Parameters map[string]string

type JsonLangData struct {
	Triggers []string `json:"triggers"`
	Results  []string `json:"results"`
}
