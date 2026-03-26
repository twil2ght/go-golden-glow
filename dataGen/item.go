package dataGen

type langData struct {
	tv       []string
	rv       []string
	params   Parameters //for rvGen
	langType LangType
}

func (l *langData) Results() []string {
	return l.rv
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

// deprecated
// use NewLangData or SNew instead
func New(triggers []string, params Parameters, langType LangType) Item {
	return &langData{
		tv:       triggers,
		params:   params,
		langType: langType,
	}
}
func NewLangData(tv []string, rv []string, params Parameters, langType LangType) Item {
	return &langData{
		tv:       tv,
		rv:       rv,
		params:   params,
		langType: langType,
	}
}

// SNew simplified version of NewLangData
func SNew(trigger, result string, params Parameters, langType LangType) Item {
	return &langData{
		tv:       []string{trigger},
		rv:       []string{result},
		params:   params,
		langType: langType,
	}
}
