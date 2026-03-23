package container

import (
	"errors"
	"goldenglow/node"
	"regexp"
)

type factory struct {
	encoder    node.Encoder
	fetcher    Fetcher
	positioner Positioner
	varReg     *regexp.Regexp
}

func (f *factory) WithVarReg(variableReg *regexp.Regexp) {
	f.varReg = variableReg
}

func (f *factory) New(hashValue string) (Item, error) {
	return New(hashValue, f.fetcher, f.encoder, f.varReg)
}
func NewFactory(fetcher Fetcher, encoder node.Encoder, p Positioner) (Factory, error) {
	if fetcher == nil {
		return nil, errors.New("fetcher is nil")
	}
	if encoder == nil {
		return nil, errors.New("encoder is nil")
	}
	if p == nil {
		return nil, errors.New("positioner is nil")
	}
	return &factory{
		fetcher:    fetcher,
		encoder:    encoder,
		positioner: p,
	}, nil
}
func (f *factory) Encoder() node.Encoder  { return f.encoder }
func (f *factory) Positioner() Positioner { return f.positioner }
