package m

type H map[string]any
type Hash Map[struct{}]
type Map[T any] map[string]T

func ToHash(tar []string) Hash {
	hash := make(Hash, len(tar))
	for _, t := range tar {
		hash[t] = struct{}{}
	}
	return hash
}
