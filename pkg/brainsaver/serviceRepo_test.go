package brainsaver

import (
	"goldenglow/m"
	"reflect"
	"testing"
)

type testRepo struct {
	data map[string]m.Hash
}

func (t *testRepo) HGet(tag string) (m.Hash, error) {
	if val, ok := t.data[tag]; ok {
		return val, nil
	}
	return nil, nil
}

func (t *testRepo) HSet(tag string, value m.Hash) error {
	if t.data == nil {
		t.data = make(map[string]m.Hash)
	}
	t.data[tag] = value
	return nil
}

func (t *testRepo) Get(key string) (string, error) {
	return "", nil
}

func (t *testRepo) Set(key, value string) error {
	return nil
}

func (t *testRepo) HDel(tag string, subKeys ...string) {
	if _, ok := t.data[tag]; !ok {
		return
	}
	if len(subKeys) == 0 {
		delete(t.data, tag)
		return
	}
	for _, subKey := range subKeys {
		delete(t.data[tag], subKey)
	}
	if len(t.data[tag]) == 0 {
		delete(t.data, tag)
	}
}

func (t *testRepo) Shutdown() error {
	return nil
}

func (t *testRepo) Init() error {
	return nil
}

func TestService_Save(t *testing.T) {
	repo := &testRepo{}
	service := NewService(repo)

	tv := m.Hash{"t1": {}, "t2": {}}
	rv := m.Hash{"r1": {}, "r2": {}}

	service.Save(tv, rv)

	// Check hash key
	hashKey := genHashForTR(tv, rv)
	if hashKey == "" {
		t.Errorf("Hash key should not be empty")
	}

	// Check TTag
	TTag := prefixC2T + hashKey
	tData, err := repo.HGet(TTag)
	if err != nil {
		t.Errorf("Error getting TTag: %v", err)
	}
	if !reflect.DeepEqual(tData, tv) {
		t.Errorf("Expected T data %v, got %v", tv, tData)
	}

	// Check RTag
	RTag := prefixC2R + hashKey
	rData, err := repo.HGet(RTag)
	if err != nil {
		t.Errorf("Error getting RTag: %v", err)
	}
	if !reflect.DeepEqual(rData, rv) {
		t.Errorf("Expected R data %v, got %v", rv, rData)
	}

	// Check node registrations
	for tKey := range tv {
		tag := prefixT2C + tKey
		cMap, err := repo.HGet(tag)
		if err != nil {
			t.Errorf("Error getting tag %s: %v", tag, err)
		}
		if cMap == nil || cMap[hashKey] != (struct{}{}) {
			t.Errorf("Expected hashKey in cMap for %s", tKey)
		}
	}

	for rKey := range rv {
		tag := prefixR2C + rKey
		cMap, err := repo.HGet(tag)
		if err != nil {
			t.Errorf("Error getting tag %s: %v", tag, err)
		}
		if cMap == nil || cMap[hashKey] != (struct{}{}) {
			t.Errorf("Expected hashKey in cMap for %s", rKey)
		}
	}

	// Check nodeSet
	nodeSet, err := repo.HGet(KeyNodeSet)
	if err != nil {
		t.Errorf("Error getting nodeSet: %v", err)
	}
	expectedNodes := m.Hash{"t1": {}, "t2": {}, "r1": {}, "r2": {}}
	if !reflect.DeepEqual(nodeSet, expectedNodes) {
		t.Errorf("Expected nodeSet %v, got %v", expectedNodes, nodeSet)
	}
}

func TestGenHashForTR(t *testing.T) {
	tv := m.Hash{"b": {}, "a": {}}
	rv := m.Hash{"2": {}, "1": {}}

	hash1 := genHashForTR(tv, rv)
	hash2 := genHashForTR(tv, rv)
	if hash1 != hash2 {
		t.Errorf("Hashes should be identical for same input")
	}

	// Different input
	rv2 := m.Hash{"3": {}}
	hash3 := genHashForTR(tv, rv2)
	if hash1 == hash3 {
		t.Errorf("Hashes should be different for different input")
	}
}

func TestSortedKeys(t *testing.T) {
	m := m.Hash{"c": {}, "a": {}, "b": {}}
	result := sortedKeys(m)
	expected := "a,b,c"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}
