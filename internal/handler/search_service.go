package handler

import (
	stdctx "context"
	"log"
	"net/http"
	"strings"

	"github.com/m1i3k0e7/distributed-search-engine/internal/indexing"
	"github.com/m1i3k0e7/distributed-search-engine/pkg/logger"
	proto "google.golang.org/protobuf/proto"

	"github.com/gin-gonic/gin"
	search_proto "github.com/m1i3k0e7/distributed-search-engine/api/proto/search"
	"github.com/m1i3k0e7/distributed-search-engine/internal/search"
	"github.com/m1i3k0e7/distributed-search-engine/internal/search/common"
	"github.com/m1i3k0e7/distributed-search-engine/internal/search/context"
	"github.com/m1i3k0e7/distributed-search-engine/pkg/preprocessing"
)

var Indexer indexing.IIndexer

func clearnKeywords(words []string) []string {
	keywords := make([]string, 0, len(words))
	for _, w := range words {
		word := strings.TrimSpace(strings.ToLower(w))
		if len(word) > 0 {
			keywords = append(keywords, word)
		}
	}
	return keywords
}

func Search(ctx *gin.Context) {
	var request common.SearchRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		log.Printf("bind request parameter failed: %s", err)
		ctx.String(http.StatusBadRequest, "invalid json")
		return
	}

	keywords := clearnKeywords(request.Keywords)
	if len(keywords) == 0 {
		ctx.String(http.StatusBadRequest, "Keywords must be non-empty")
		return
	}
	
	query := new(search_proto.TermQuery)
	if len(keywords) > 0 {
		for _, word := range keywords {
			query = query.And(search_proto.NewTermQuery("content", word))
		}
	}

	orFlags := []uint64{common.GetClassBits(request.Classes)}
	docs := Indexer.Search(query, 0, 0, orFlags)
	products := make([]search_proto.Product, 0, len(docs))
	for _, doc := range docs {
		var product search_proto.Product
		if err := proto.Unmarshal(doc.Bytes, &product); err == nil {
			if product.DiscountPrice >= float64(request.PriceTo) && (request.PriceTo <= 0 || product.DiscountPrice <= float64(request.PriceTo)) {
				products = append(products, product)
			}
		}
	}

	logger.Log.Printf("return %d products", len(products))
	ctx.JSON(http.StatusOK, products)
}

func SearchAll(ctx *gin.Context) {
	var request common.SearchRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		log.Printf("bind request parameter failed: %s", err)
		ctx.String(http.StatusBadRequest, "invalid json")
		return
	}

	// TODO: parse query into keywords
	keywords := preprocessing.Preprocess(request.Query)
	request.Keywords = clearnKeywords(keywords)
	if len(request.Keywords) == 0 {
		ctx.String(http.StatusBadRequest, "Query must be non-empty and contain valid keywords")
		return
	}

	searchCtx := &context.ProductSearchContext{
		Ctx:     stdctx.Background(),
		Request: &request,
		Indexer: Indexer,
	}
	searcher := search.NewAllProductSearcher()
	products := searcher.Search(searchCtx)

	ctx.JSON(http.StatusOK, products)
}