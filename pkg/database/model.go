package database

import "goldenglow/m"

type Repository interface {
	HGet(tag string) (m.Hash, error)
	HSet(tag string, value m.Hash) error
	HDel(tag string, subKeys ...string)
	Shutdown() error
	Init() error
}
type RedisRepository interface {
	Set(key, value, expiration string) error
	HGet(tag string) (m.Hash, error)
	Del(key string)
	HDel(key string, subKeys ...string)
	Shutdown() error
	Init() error
}
