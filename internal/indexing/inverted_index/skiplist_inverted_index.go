package inverted_index

import (
	"runtime"
	"sync"

	search_proto "github.com/m1i3k0e7/distributed-search-engine/api/proto/search"
	"github.com/m1i3k0e7/distributed-search-engine/pkg/concurrent"
	"github.com/huandu/skiplist"
	farmhash "github.com/leemcloughlin/gofarmhash"
	// "github.com/m1i3k0e7/distributed-search-engine/pkg/logger"
)

// the value of the inverted index is a skiplist
type SkipListReverseIndex struct {
	table *concurrent.ConcurrentHashMap // segmented hashmap, key is keyword, value is *skiplist.SkipList
	locks []sync.RWMutex          // locks for each segment, the same keys must compete for the same lock when writing
}

func NewSkipListReverseIndex(DocNumEstimate int) *SkipListReverseIndex {
	indexer := new(SkipListReverseIndex)
	indexer.table = concurrent.NewConcurrentHashMap(runtime.NumCPU(), DocNumEstimate)
	indexer.locks = make([]sync.RWMutex, 1000)
	return indexer
}

func (indexer SkipListReverseIndex) getLock(key string) *sync.RWMutex {
	n := int(farmhash.Hash32WithSeed([]byte(key), 0))
	return &indexer.locks[n%len(indexer.locks)]
}

// SkipListKey is Document.IntId, and SkipListValue is (Document.Id, Document.BitsFeature）
type SkipListValue struct {
	Id          string
	BitsFeature uint64
}

func (indexer *SkipListReverseIndex) Add(doc search_proto.Document) {
	for _, keyword := range doc.Keywords {
		key := keyword.ToString()
		lock := indexer.getLock(key)
		lock.Lock() // lock for writing
		sklValue := SkipListValue{doc.Id, doc.BitsFeature}
		if value, exists := indexer.table.Get(key); exists {
			list := value.(*skiplist.SkipList)
			list.Set(doc.IntId, sklValue)
		} else {
			list := skiplist.New(skiplist.Uint64)
			list.Set(doc.IntId, sklValue)
			indexer.table.Set(key, list)
		}
		// logger.Log.Printf("add key %s value %d to reverse index\n", key, doc.IntId)
		lock.Unlock()
	}
}

func (indexer *SkipListReverseIndex) Delete(IntId uint64, keyword *search_proto.Keyword) {
	key := keyword.ToString()
	lock := indexer.getLock(key)
	lock.Lock() // lock for deleting
	if value, exists := indexer.table.Get(key); exists {
		list := value.(*skiplist.SkipList)
		list.Remove(IntId)
	}
	lock.Unlock()
}

func IntersectionOfSkipList(lists ...*skiplist.SkipList) *skiplist.SkipList {
	if len(lists) == 0 {
		return nil
	}

	if len(lists) == 1 {
		return lists[0]
	}

	result := skiplist.New(skiplist.Uint64)
	currNodes := make([]*skiplist.Element, len(lists)) // assign a pointer for each skiplist
	for i, list := range lists {
		if list == nil || list.Len() == 0 { // there is an empty skiplist, so the intersection must be empty
			return nil
		}
		currNodes[i] = list.Front()
	}
	for {
		maxList := make(map[int]struct{}, len(currNodes)) // map to store the indices of skiplists that have the current maximum value
		var maxValue uint64 = 0
		for i, node := range currNodes {
			if node.Key().(uint64) > maxValue {
				maxValue = node.Key().(uint64)
				maxList = map[int]struct{}{i: {}} // update maxList to only contain the current index
			} else if node.Key().(uint64) == maxValue {
				maxList[i] = struct{}{}
			}
		}

		if len(maxList) == len(currNodes) { // all skiplists have the same current value, so it's part of the intersection
			result.Set(currNodes[0].Key(), currNodes[0].Value)
			for i, node := range currNodes { // move all pointers forward
				currNodes[i] = node.Next()
				if currNodes[i] == nil {
					return result
				}
			}
		} else {
			for i, node := range currNodes {
				if _, exists := maxList[i]; !exists {
					currNodes[i] = node.Next()
					if currNodes[i] == nil { // if any pointer reaches the end, the intersection is complete
						return result
					}
				}
			}
		}
	}
}

func UnionsetOfSkipList(lists ...*skiplist.SkipList) *skiplist.SkipList {
	if len(lists) == 0 {
		return nil
	}

	if len(lists) == 1 {
		return lists[0]
	}

	result := skiplist.New(skiplist.Uint64)
	for _, list := range lists {
		if list == nil {
			continue
		}

		node := list.Front()
		for node != nil {
			result.Set(node.Key(), node.Value)
			node = node.Next()
		}
	}
	return result
}

func (indexer SkipListReverseIndex) FilterByBits(bits uint64, onFlag uint64, offFlag uint64, orFlags []uint64) bool {
	// onFlag must be fully matched
	if bits & onFlag != onFlag {
		return false
	}

	// offFlag must be fully unmatched
	if bits & offFlag != 0 {
		return false
	}

	// orFlags, any one of them can be matched
	for _, orFlag := range orFlags {
		if (orFlag > 0) && (bits & orFlag <= 0) { // orFlag must be at least partially matched
			return false
		}
	}

	return true
}

func (indexer SkipListReverseIndex) search(q *search_proto.TermQuery, onFlag uint64, offFlag uint64, orFlags []uint64) *skiplist.SkipList {
	if q.Keyword != nil {
		Keyword := q.Keyword.ToString()
		lock := indexer.getLock(Keyword)
		lock.RLock() // read lock for searching
		defer lock.RUnlock()
		if value, exists := indexer.table.Get(Keyword); exists {
			result := skiplist.New(skiplist.Uint64)
			list := value.(*skiplist.SkipList)
			// util.Log.Printf("retrive %d docs by key %s", list.Len(), Keyword)
			node := list.Front()
			for node != nil {
				intId := node.Key().(uint64)
				skv, _ := node.Value.(SkipListValue)
				flag := skv.BitsFeature
				if intId > 0 && indexer.FilterByBits(flag, onFlag, offFlag, orFlags) {
					result.Set(intId, skv)
				}
				node = node.Next()
			}

			return result
		}
	} else if len(q.Must) > 0 {
		results := make([]*skiplist.SkipList, 0, len(q.Must))
		for _, q := range q.Must {
			results = append(results, indexer.search(q, onFlag, offFlag, orFlags))
		}

		return IntersectionOfSkipList(results...)
	} else if len(q.Should) > 0 {
		results := make([]*skiplist.SkipList, 0, len(q.Should))
		for _, q := range q.Should {
			results = append(results, indexer.search(q, onFlag, offFlag, orFlags))
		}

		return UnionsetOfSkipList(results...)
	}

	return nil
}

func (indexer SkipListReverseIndex) Search(query *search_proto.TermQuery, onFlag uint64, offFlag uint64, orFlags []uint64) []string {
	result := indexer.search(query, onFlag, offFlag, orFlags)
	if result == nil {
		return nil
	}

	arr := make([]string, 0, result.Len())
	node := result.Front()
	for node != nil {
		skv, _ := node.Value.(SkipListValue)
		arr = append(arr, skv.Id)
		node = node.Next()
	}
	
	return arr
}
