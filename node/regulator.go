package node

import (
	"fmt"
	"goldenglow/storage"
	"strings"
)

type regulator struct {
	db storage.KVLite
}

func (r *regulator) Do(str string) string {
	res := make([]string, 0)
	for t := range strings.FieldsSeq(str) {
		if key, err := r.db.Get(t); err == nil {
			res = append(res, key)
		} else {
			res = append(res, t)
		}
	}
	return strings.Join(res, " ")
}
func NewRegulator(db storage.KVLite) (Regulator, error) {
	if db == nil {
		return nil, fmt.Errorf("regulator init: db is nil")
	}
	return &regulator{
		db: db,
	}, nil
}
