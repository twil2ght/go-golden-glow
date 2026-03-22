package storage

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
