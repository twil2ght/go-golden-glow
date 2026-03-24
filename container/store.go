package container

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"goldenglow/m"
	"goldenglow/node"
	"goldenglow/pkg/log"
	"sort"
	"strings"
)

type store struct {
	db      Repository
	encoder node.Encoder
}

func (s *store) Save(tv, rv m.Hash) error {
	head := "store"

	hashKey := s.generateHash(tv, rv)

	if e, _ := s.db.HGet(hashKey); e != nil {
		//return fmt.Errorf("%s: %s already exists", head, hashKey)
		return nil
	}

	TTag := prefixC2T + hashKey
	RTag := prefixC2R + hashKey

	if err := s.db.HSet(TTag, tv); err != nil {
		return fmt.Errorf("%s: %w", head, log.NotFound(TTag))
	}
	if err := s.db.HSet(RTag, rv); err != nil {
		return fmt.Errorf("%s: %w", head, log.NotFound(RTag))
	}

	for t := range tv {
		if err := s.nodeRegister(t, TypeT, hashKey); err != nil {
			return fmt.Errorf("%s: %w", head, err)
		}
	}
	for r := range rv {
		if err := s.nodeRegister(r, TypeR, hashKey); err != nil {
			return fmt.Errorf("%s: %w", head, err)
		}
	}

	return nil
}

func (s *store) nodeRegister(value, kind string, hashValue string) error {
	encodedValue := s.encoder.Do(value)
	var tag string
	switch kind {
	case TypeR:
		tag = prefixR2C + encodedValue
	case TypeT:
		tag = prefixT2C + encodedValue
	}

	// 读取已有的容器ID集合
	cMap, err := s.db.HGet(tag)
	if err != nil {
		return log.NotFound(tag)
	}

	// 初始化 map（防止 nil）
	if cMap == nil {
		cMap = make(m.Hash)
	}

	// 添加当前容器ID
	cMap[hashValue] = struct{}{}

	// 写回数据库
	if err := s.db.HSet(tag, cMap); err != nil {
		return err
	}

	return nil
}

func (s *store) generateHash(tv, rv map[string]struct{}) string {
	// 排序 key（map 遍历无序，必须排序才能保证相同内容 hash 一致）
	sortedT := sortedKeys(tv)
	sortedR := sortedKeys(rv)

	// 拼接成唯一字符串
	str := fmt.Sprintf("T:[%s]|R:[%s]", sortedT, sortedR)

	// MD5 哈希
	hash := md5.Sum([]byte(str))
	return hex.EncodeToString(hash[:])
}

func sortedKeys(m map[string]struct{}) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return strings.Join(keys, ",")
}

func NewStore(db Repository, encoder node.Encoder) Store {
	return &store{
		db:      db,
		encoder: encoder,
	}
}
