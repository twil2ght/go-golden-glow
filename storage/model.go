package storage

import "goldenglow/m"

type KVLite interface {
	LightRepository
	Keyof(value string) (string, error)
}
type Repository interface {
	HashRepository
	LightRepository
	Shutdown() error
	Init() error
}
type HashRepository interface {
	HGet(tag string) (m.Hash, error)
	HSet(tag string, value m.Hash) error
}
type LightRepository interface {
	Get(key string) (string, error)
	Set(key, value string) error
}
