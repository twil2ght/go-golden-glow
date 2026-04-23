package brainsaver

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"goldenglow/m"
	"goldenglow/pkg/database"
	"sort"
	"strings"
)

var (
	prefixC2T  = "C->T:" // Container -> T Node
	prefixC2R  = "C->R:" // Container -> R Node
	prefixT2C  = "T->C:" // T Node -> Container
	prefixR2C  = "R->C:" // R Node -> Container
	TypeT      = "T"
	TypeR      = "R"
	KeyNodeSet = "nodeSet"
)

type Service interface {
	Save(t, r m.Hash)
}

type service struct {
	repo database.Repository
}

func (s *service) Save(tv, rv m.Hash) {
	hashKey := genHashForTR(tv, rv)

	if e, _ := s.repo.HGet(hashKey); e != nil {
		return
	}

	TTag := prefixC2T + hashKey
	RTag := prefixC2R + hashKey

	if err := s.repo.HSet(TTag, tv); err != nil {
		return
	}
	if err := s.repo.HSet(RTag, rv); err != nil {
		return
	}

	for t := range tv {
		if err := s.nodeRegister(t, TypeT, hashKey); err != nil {
			return
		}
	}
	for r := range rv {
		if err := s.nodeRegister(r, TypeR, hashKey); err != nil {
			return
		}
	}

	nodeSet, _ := s.repo.HGet(KeyNodeSet)
	if nodeSet == nil {
		nodeSet = make(map[string]struct{})
	}

	for nv := range tv {
		nodeSet[nv] = struct{}{}
	}
	for nv := range rv {
		nodeSet[nv] = struct{}{}
	}
	_ = s.repo.HSet(KeyNodeSet, nodeSet)
}

func (s *service) nodeRegister(value, kind string, hashValue string) error {
	var tag string
	switch kind {
	case TypeR:
		tag = prefixR2C + value
	case TypeT:
		tag = prefixT2C + value
	}

	cMap, _ := s.repo.HGet(tag)

	if cMap == nil {
		cMap = make(m.Hash)
	}

	cMap[hashValue] = struct{}{}

	if err := s.repo.HSet(tag, cMap); err != nil {
		return err
	}

	return nil
}

func genHashForTR(tv, rv m.Hash) string {
	sortedT := sortedKeys(tv)
	sortedR := sortedKeys(rv)
	str := fmt.Sprintf("T:[%s]|R:[%s]", sortedT, sortedR)
	hash := md5.Sum([]byte(str))
	return hex.EncodeToString(hash[:])
}

func sortedKeys(m m.Hash) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return strings.Join(keys, ",")
}
func NewService(repo database.Repository) Service {
	return &service{repo: repo}
}
func DefaultService() Service {
	return NewService(database.DefaultJSONRepo())
}
