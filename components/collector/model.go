package collector

type Instance interface {
	SetSource(userTag string, selfTag string)
	Save() error
	Run()
}
