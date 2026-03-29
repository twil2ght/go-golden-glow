package dataGen

type langData struct {
	tv       []string
	rv       []string
	params   Parameters //for rvGen
	langType LangType
}

func (l *langData) Results() []string {
	return l.removeNull(l.rv)
}

func (l *langData) Triggers() []string {
	return l.removeNull(l.tv)
}
func (l *langData) removeNull(nv []string) []string {
	var result []string
	for _, v := range nv {
		if v != "" {
			result = append(result, v)
		}
	}
	return result
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
