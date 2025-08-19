package indexing

import (
	"bytes"
	"encoding/gob"
	"strings"
	"sync/atomic"

	"github.com/m1i3k0e7/distributed-search-engine/internal/indexing/kvdb"
	inverted_index "github.com/m1i3k0e7/distributed-search-engine/internal/indexing/inverted_index"
	search_proto "github.com/m1i3k0e7/distributed-search-engine/api/proto/search"
	"github.com/m1i3k0e7/distributed-search-engine/pkg/logger"
)

// enclosed by Indexer, which is the main interface for indexing operations.
type Indexer struct {
	forwardIndex kvdb.IKeyValueDB
	reverseIndex inverted_index.IReverseIndexer
	maxIntId     uint64
}

func (indexer *Indexer) Init(DocNumEstimate int, dbtype int, DataDir string) error {
	db, err := kvdb.GetKvDb(dbtype, DataDir) // get the key-value database
	if err != nil {
		return err
	}

	indexer.forwardIndex = db
	indexer.reverseIndex = inverted_index.NewSkipListReverseIndex(DocNumEstimate)

	return nil
}

func (indexer *Indexer) LoadFromIndexFile() int {
	reader := bytes.NewReader([]byte{})
	n := indexer.forwardIndex.IterDB(func(k, v []byte) error {
		reader.Reset(v)
		decoder := gob.NewDecoder(reader)
		var doc search_proto.Document
		err := decoder.Decode(&doc)
		if err != nil {
			logger.Log.Printf("gob decode document failed：%s", err)
			return nil
		}

		indexer.reverseIndex.Add(doc)

		return err
	})
	
	logger.Log.Printf("load %d data from forward index %s", n, indexer.forwardIndex.GetDbPath())

	return int(n)
}

func (indexer *Indexer) Close() error {
	return indexer.forwardIndex.Close()
}

func (indexer *Indexer) AddDoc(doc search_proto.Document) (int, error) {
	docId := strings.TrimSpace(doc.Id)
	if len(docId) == 0 {
		return 0, nil
	}

	doc.IntId = atomic.AddUint64(&indexer.maxIntId, 1) // assign a new IntId to the document
	// write the document to the kvdb
	var value bytes.Buffer
	encoder := gob.NewEncoder(&value)
	if err := encoder.Encode(doc); err == nil {
		indexer.forwardIndex.Set([]byte(docId), value.Bytes())
	} else {
		return 0, err
	}

	// add the document to the inverted index
	indexer.reverseIndex.Add(doc)
	return 1, nil
}

func (indexer *Indexer) UpdateDoc(doc search_proto.Document) (int, error) {
	docId := strings.TrimSpace(doc.Id)
	if len(docId) == 0 {
		return 0, nil
	}
	
	indexer.DeleteDoc(docId)
	return indexer.AddDoc(doc)
}

func (indexer *Indexer) DeleteDoc(docId string) int {
	forwardKey := []byte(docId)
	docBs, err := indexer.forwardIndex.Get(forwardKey)
	if err == nil {
		reader := bytes.NewReader([]byte{})
		if len(docBs) > 0 {
			reader.Reset(docBs)
			decoder := gob.NewDecoder(reader)
			var doc search_proto.Document
			err := decoder.Decode(&doc)
			if err == nil {
				for _, kw := range doc.Keywords {
					indexer.reverseIndex.Delete(doc.IntId, kw)
				}
			}
		}
	} else {
		return 0
	}

	// Delete the document from the forward index
	if err := indexer.forwardIndex.Delete(forwardKey); err != nil {
		return 0
	}

	return 1
}

func (indexer *Indexer) Search(query *search_proto.TermQuery, onFlag uint64, offFlag uint64, orFlags []uint64) []*search_proto.Document {
	docIds := indexer.reverseIndex.Search(query, onFlag, offFlag, orFlags)
	if len(docIds) == 0 {
		logger.Log.Printf("no documents found for query: %s", query)
		return nil
	}

	keys := make([][]byte, 0, len(docIds))
	for _, docId := range docIds {
		keys = append(keys, []byte(docId))
	}

	docs, err := indexer.forwardIndex.BatchGet(keys)
	if err != nil {
		logger.Log.Printf("read kvdb failed: %s", err)
		return nil
	}

	result := make([]*search_proto.Document, 0, len(docs))
	reader := bytes.NewReader([]byte{})
	for _, docBs := range docs {
		if len(docBs) > 0 {
			reader.Reset(docBs)
			decoder := gob.NewDecoder(reader)
			var doc search_proto.Document
			err := decoder.Decode(&doc)
			if err == nil {
				result = append(result, &doc)
			}
		}
	}

	return result
}

func (indexer *Indexer) Count() int {
	n := 0
	indexer.forwardIndex.IterKey(func(k []byte) error {
		n++
		return nil
	})

	return n
}
