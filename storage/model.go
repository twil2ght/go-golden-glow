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
}
type HashRepository interface {
	HGet(tag string) (m.Hash, error)
	HSet(tag string, value m.Hash) error
}
type LightRepository interface {
	Get(key string) (string, error)
	Set(key, value string) error
}
type kvLite struct {
	lr LightRepository
}

const (
	prefixKeyOf = "KeyOf:"
)

func (kv *kvLite) Get(key string) (string, error) {
	return kv.lr.Get(key)
}
func (kv *kvLite) Set(key, value string) error {
	// 存 key → value
	if err := kv.lr.Set(key, value); err != nil {
		return err
	}
	// 存 value → key（反向映射）
	return kv.lr.Set(prefixKeyOf+value, key)
}
func (kv *kvLite) Keyof(value string) (string, error) {
	return kv.Get(prefixKeyOf + value)
}
func NewKVLite(repository LightRepository) KVLite {
	return &kvLite{
		lr: repository,
	}
}
