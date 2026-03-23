package dataGen

type langData struct {
	tv       []string
	params   Parameters //for rvGen
	langType LangType
}

func (l *langData) Triggers() []string {
	return l.tv
}

func (l *langData) LangType() LangType {
	return l.langType
}

func (l *langData) Params() Parameters {
	return l.params
}
func New(triggers []string, params Parameters, langType LangType) Item {
	return &langData{
		tv:       triggers,
		params:   params,
		langType: langType,
	}
}
