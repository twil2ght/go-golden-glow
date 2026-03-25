package container

import (
	"goldenglow/node"
	"goldenglow/utils"
	"regexp"
)

type factory struct {
	fetcher    Fetcher
	positioner Positioner
	varReg     *regexp.Regexp
}

func (f *factory) WithVarReg(variableReg *regexp.Regexp) {
	f.varReg = variableReg
}

func (f *factory) New(hashValue string) (Item, error) {
	return New(hashValue, f.fetcher, f.Encoder(), f.varReg)
}
func NewFactory(fetcher Fetcher, p Positioner) (Factory, error) {
	if err := utils.NotNull("fetcher", fetcher, "positioner", p); err != nil {
		return nil, err
	}
	return &factory{
		fetcher:    fetcher,
		positioner: p,
	}, nil
}
func (f *factory) Encoder() node.Encoder  { return f.positioner.Encoder() }
func (f *factory) Positioner() Positioner { return f.positioner }
