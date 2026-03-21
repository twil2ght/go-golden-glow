package storage

import (
	"database/sql"
	"goldenglow/m"
)

// DBRepository TODO implement
type DBRepository struct {
	db *sql.DB
}

func (D *DBRepository) HGet(tag string) (m.Hash, error) {
	//TODO implement me
	panic("implement me")
}

func (D *DBRepository) HSet(tag string, value m.Hash) error {
	//TODO implement me
	panic("implement me")
}
