package container

import (
	"errors"
	"goldenglow/node"
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
	if fetcher == nil {
		return nil, errors.New("fetcher is nil")
	}
	if p == nil {
		return nil, errors.New("positioner is nil")
	}
	return &factory{
		fetcher:    fetcher,
		positioner: p,
	}, nil
}
func (f *factory) Encoder() node.Encoder  { return f.positioner.Encoder() }
func (f *factory) Positioner() Positioner { return f.positioner }
