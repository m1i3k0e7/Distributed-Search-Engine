package ranking

import (
	"github.com/m1i3k0e7/distributed-search-engine/pkg/preprocessing"
)

func preprocess(text string) []string {
	// This function can be customized to include more advanced preprocessing
	return preprocessing.PreprocessForLargeDataset(text)
}