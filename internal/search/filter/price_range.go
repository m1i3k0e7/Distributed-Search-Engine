package filter

import (
	search_proto "github.com/m1i3k0e7/distributed-search-engine/api/proto/search"
	"github.com/m1i3k0e7/distributed-search-engine/internal/search/context"
)

type ViewFilter struct {
}

func (ViewFilter) Apply(ctx *context.ProductSearchContext) {
	request := ctx.Request
	if request == nil {
		return
	}
	if request.PriceFrom >= request.PriceTo {
		return
	}

	products := make([]*search_proto.Product, 0, len(ctx.Products))
	for _, product := range ctx.Products {
		if product.DiscountPrice >= float64(request.PriceFrom) && product.DiscountPrice <= float64(request.PriceTo) {
			products = append(products, product)
		}
	}
	
	ctx.Products = products
}
