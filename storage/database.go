package storage

import (
	"database/sql"
	"goldenglow/m"
)

// dbRepository TODO implement
type dbRepository struct {
	db *sql.DB
}

func (D *dbRepository) Get(key string) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (D *dbRepository) Set(key, value string) error {
	//TODO implement me
	panic("implement me")
}

func (D *dbRepository) Shutdown() error {
	//TODO implement me
	panic("implement me")
}

func (D *dbRepository) HGet(tag string) (m.Hash, error) {
	//TODO implement me
	panic("implement me")
}

func (D *dbRepository) HSet(tag string, value m.Hash) error {
	//TODO implement me
	panic("implement me")
}
func NewDBRepository(db *sql.DB) Repository {
	return &dbRepository{db: db}
}
