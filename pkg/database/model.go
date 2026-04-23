package database

import "goldenglow/m"

type Repository interface {
	HGet(tag string) (m.Hash, error)
	HSet(tag string, value m.Hash) error
	Shutdown() error
	Init() error
}
type RedisRepository interface {
	Set(key, value, expiration string) error
	HGet(tag string) (m.Hash, error)
	Shutdown() error
	Init() error
}
