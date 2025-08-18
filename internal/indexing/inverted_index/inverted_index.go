package inverted_index

import search_proto "github.com/m1i3k0e7/distributed-search-engine/api/proto/search"

type IReverseIndexer interface {
	Add(doc search_proto.Document)                                                              // Add a doc to the index
	Delete(IntId uint64, keyword *search_proto.Keyword)                                         // Delete a doc from the index
	Search(q *search_proto.TermQuery, onFlag uint64, offFlag uint64, orFlags []uint64) []string // Search the index with a term query
}
