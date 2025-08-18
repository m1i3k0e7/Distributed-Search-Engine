package context

import (
	"context"

	search_proto "github.com/m1i3k0e7/distributed-search-engine/api/proto/search"
	"github.com/m1i3k0e7/distributed-search-engine/internal/search/common"
	"github.com/m1i3k0e7/distributed-search-engine/internal/indexing"
)

type ProductSearchContext struct {
	Ctx     context.Context
	Indexer indexing.IIndexer
	Request *common.SearchRequest
	Products  []*search_proto.Product
}

type Filter interface {
	Apply(*ProductSearchContext)
}