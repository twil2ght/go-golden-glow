package dataGen

type Item interface {
	Triggers() []string
	LangType() LangType
	Params() Parameters
}
type Generator interface {
	Register(name string, langItem Item)
	Run() error
}
type LangType string
type Parameters map[string]string

const (
	langTypeDefault LangType = "default"
	langTypeCheck   LangType = "check"
	langTypeExtract LangType = "extract"
	RootDir                  = "../../data/"
	jsonExt                  = ".json"
	KeyDefault               = "[node]"
)

type JsonLangData struct {
	Triggers []string `json:"triggers"`
	Results  []string `json:"results"`
}
