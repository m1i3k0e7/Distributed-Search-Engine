package recaller

import (
	search_proto "github.com/m1i3k0e7/distributed-search-engine/api/proto/search"
	"github.com/m1i3k0e7/distributed-search-engine/internal/search/context"
	"github.com/m1i3k0e7/distributed-search-engine/internal/search/common"
	proto "google.golang.org/protobuf/proto"
)

type KeywordRecaller struct {
}

func (KeywordRecaller) Recall(ctx *context.ProductSearchContext) []*search_proto.Product {
	request := ctx.Request
	if request == nil {
		return nil
	}

	indexer := ctx.Indexer
	if indexer == nil {
		return nil
	}

	keywords := request.Keywords
	query := new(search_proto.TermQuery)
	if len(keywords) > 0 {
		for _, word := range keywords {
			query = query.And(search_proto.NewTermQuery("content", word))
		}
	}

	orFlags := []uint64{common.GetClassBits(request.Classes)}
	docs := indexer.Search(query, 0, 0, orFlags)
	products := make([]*search_proto.Product, 0, len(docs))
	for _, doc := range docs {
		var product search_proto.Product
		if err := proto.Unmarshal(doc.Bytes, &product); err == nil {
			products = append(products, &product)
		}
	}

	return products
}
