package node

import (
	"fmt"
	"strings"
)

type regulator struct {
	repo LightRepository
}

func (r *regulator) Do(str string) string {
	res := make([]string, 0)
	for t := range strings.FieldsSeq(str) {
		if key, err := r.repo.Get(t); err == nil {
			res = append(res, key)
		} else {
			res = append(res, t)
		}
	}
	return strings.Join(res, " ")
}
func NewRegulator(repo LightRepository) (Regulator, error) {
	if repo == nil {
		return nil, fmt.Errorf("regulator init: repo is nil")
	}
	return &regulator{
		repo: repo,
	}, nil
}
