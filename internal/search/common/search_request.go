package common

type SearchRequest struct {
	Classes  []string
	Keywords []string
	Query	string
	PriceFrom int
	PriceTo   int
}
