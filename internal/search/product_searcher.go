package search

import (
	"reflect"
	"sync"
	"time"

	search_proto "github.com/m1i3k0e7/distributed-search-engine/api/proto/search"
	"github.com/m1i3k0e7/distributed-search-engine/internal/search/filter"
	"github.com/m1i3k0e7/distributed-search-engine/internal/search/recaller"
	"github.com/m1i3k0e7/distributed-search-engine/pkg/logger"
	"golang.org/x/exp/maps"
	"github.com/m1i3k0e7/distributed-search-engine/internal/search/context"
)

type Recaller interface {
	Recall(*context.ProductSearchContext) []*search_proto.Product
}

type ProductSearcher struct {
	Recallers []Recaller
	Filters   []context.Filter // Use context.Filter
}

func (searcher *ProductSearcher) WithRecaller(recaller ...Recaller) {
	searcher.Recallers = append(searcher.Recallers, recaller...)
}

func (searcher *ProductSearcher) WithFilter(filter ...context.Filter) {
	searcher.Filters = append(searcher.Filters, filter...)
}

func (searcher *ProductSearcher) Recall(searchContext *context.ProductSearchContext) {
	if len(searcher.Recallers) == 0 {
		return
	}

	// parallel recall
	collection := make(chan *search_proto.Product, 1000)
	wg := sync.WaitGroup{}
	wg.Add(len(searcher.Recallers))
	for _, recaller := range searcher.Recallers {
		go func(recaller Recaller) {
			defer wg.Done()
			rule := reflect.TypeOf(recaller).Name()
			result := recaller.Recall(searchContext)
			logger.Log.Printf("recall %d docs by %s", len(result), rule)
			for _, product := range result {
				collection <- product
			}
		}(recaller)
	}

	// merge results, deduplicate by product ID
	productMap := make(map[string]*search_proto.Product, 1000)
	receiveFinish := make(chan struct{})
	go func() {
		for {
			product, ok := <-collection
			if !ok {
				break
			}
			productMap[product.Id] = product
		}
		receiveFinish <- struct{}{}
	}()
	wg.Wait()
	close(collection)
	<-receiveFinish

	searchContext.Products = maps.Values(productMap)
}

func (searcher *ProductSearcher) Filter(searchContext *context.ProductSearchContext) {
	// apply each filter in order
	for _, filter := range searcher.Filters {
		filter.Apply(searchContext)
	}
}

func (searcher *ProductSearcher) Search(searchContext *context.ProductSearchContext) []*search_proto.Product {
	t1 := time.Now()

	searcher.Recall(searchContext)
	t2 := time.Now()
	logger.Log.Printf("recall %d docs in %d ms", len(searchContext.Products), t2.Sub(t1).Milliseconds())

	searcher.Filter(searchContext)
	t3 := time.Now()
	logger.Log.Printf("after filter remain %d docs in %d ms", len(searchContext.Products), t3.Sub(t2).Milliseconds())

	return searchContext.Products
}

type AllProductSearcher struct {
	ProductSearcher
}

func NewAllProductSearcher() *AllProductSearcher {
	searcher := new(AllProductSearcher)
	searcher.WithRecaller(recaller.KeywordRecaller{})
	searcher.WithFilter(filter.ViewFilter{})
	
	return searcher
}