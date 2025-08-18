package concurrent

import (
	"sync"
	farmhash "github.com/leemcloughlin/gofarmhash"
	"golang.org/x/exp/maps"
)

type ConcurrentHashMap struct {
	mps   []map[string]any // multiple small maps
	seg   int              // number of small maps
	locks []sync.RWMutex   // each small map has its own lock
	seed  uint32           // hash seed
}

func NewConcurrentHashMap(seg, cap int) *ConcurrentHashMap {
	mps := make([]map[string]any, seg)
	locks := make([]sync.RWMutex, seg)
	for i := 0; i < seg; i++ {
		mps[i] = make(map[string]any, cap/seg)
		// locks[i] = sync.RWMutex{}
	}

	return &ConcurrentHashMap{
		mps:   mps,
		seg:   seg,
		seed:  0,
		locks: locks,
	}
}

func (m *ConcurrentHashMap) getSegIndex(key string) int {
	hash := int(farmhash.Hash32WithSeed([]byte(key), m.seed))
	return hash % m.seg
}

func (m *ConcurrentHashMap) Set(key string, value any) {
	index := m.getSegIndex(key)
	m.locks[index].Lock()
	defer m.locks[index].Unlock()
	m.mps[index][key] = value
}

func (m *ConcurrentHashMap) Get(key string) (any, bool) {
	index := m.getSegIndex(key)
	m.locks[index].RLock()
	defer m.locks[index].RUnlock()

	value, exists := m.mps[index][key]

	return value, exists
}

func (m *ConcurrentHashMap) CreateIterator() *ConcurrentHashMapIterator {
	keys := make([][]string, 0, len(m.mps))
	for _, mp := range m.mps {
		row := maps.Keys(mp)
		keys = append(keys, row)
	}

	return &ConcurrentHashMapIterator{
		cm:       m,
		keys:     keys,
		rowIndex: 0,
		colIndex: 0,
	}
}

type MapEntry struct {
	Key   string
	Value any
}

type MapIterator interface {
	Next() *MapEntry
}

type ConcurrentHashMapIterator struct {
	cm       *ConcurrentHashMap
	keys     [][]string
	rowIndex int
	colIndex int
}

func (iter *ConcurrentHashMapIterator) Next() *MapEntry {
	if iter.rowIndex >= len(iter.keys) {
		return nil
	}

	row := iter.keys[iter.rowIndex]
	if len(row) == 0 {
		iter.rowIndex += 1
		return iter.Next()
	}

	key := row[iter.colIndex]
	value, _ := iter.cm.Get(key)
	if iter.colIndex >= len(row)-1 {
		iter.rowIndex += 1
		iter.colIndex = 0
	} else {
		iter.colIndex += 1
	}
	
	return &MapEntry{key, value}
}
