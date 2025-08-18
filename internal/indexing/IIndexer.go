package indexing

import search_proto "github.com/m1i3k0e7/distributed-search-engine/api/proto/search"

type IIndexer interface {
	AddDoc(doc search_proto.Document) (int, error)
	UpdateDoc(doc search_proto.Document) (int, error)
	DeleteDoc(docId string) int
	Search(query *search_proto.TermQuery, onFlag uint64, offFlag uint64, orFlags []uint64) []*search_proto.Document
	Count() int
	Close() error
}
