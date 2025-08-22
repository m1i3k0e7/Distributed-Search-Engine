package ranking

import (
	"math"
	"sort"

	search_proto "github.com/m1i3k0e7/distributed-search-engine/api/proto/search"
)

const (
	bm25K1 = 1.2 // K1 parameter for BM25
	bm25B  = 0.75 // B parameter for BM25
)

func RankDocumentByBM25(query string, docs []*search_proto.Product) []*search_proto.Product {
	if len(docs) == 0 || query == "" {
		return docs
	}

	// Step 1: Preprocess documents and calculate term frequencies and document lengths.
	processedDocs := make(map[string][]string)
	docTermCounts := make(map[string]map[string]int)
	var totalDocLength float64

	for _, doc := range docs {
		tokens := preprocess(doc.Name)
		processedDocs[doc.Id] = tokens
		totalDocLength += float64(len(tokens))

		counts := make(map[string]int)
		for _, token := range tokens {
			counts[token]++
		}
		docTermCounts[doc.Id] = counts
	}

	avgDocLength := totalDocLength / float64(len(docs))

	// Step 2: Preprocess query and calculate IDF for each unique query term.
	processedQuery := preprocess(query)
	uniqueQueryTerms := make(map[string]bool)
	for _, term := range processedQuery {
		uniqueQueryTerms[term] = true
	}

	// Calculate document frequency (n(q)) for each unique query term.
	docFreqs := make(map[string]int)
	for term := range uniqueQueryTerms {
		count := 0
		for _, counts := range docTermCounts {
			if _, exists := counts[term]; exists {
				count++
			}
		}
		docFreqs[term] = count
	}

	idf := calculateIDF(docFreqs, len(docs))

	// Step 3: Calculate BM25 score for each document.
	scores := make(map[string]float64)
	for _, doc := range docs {
		score := 0.0
		docLen := float64(len(processedDocs[doc.Id]))
		K := bm25K1 * (1 - bm25B + bm25B * docLen / avgDocLength)
		for term := range uniqueQueryTerms {
			if idfValue, ok := idf[term]; ok {
				termFreq := float64(docTermCounts[doc.Id][term])
				score += idfValue * (termFreq * (bm25K1 + 1)) / (termFreq + K)
			}
		}
		scores[doc.Id] = score
	}

	// Step 4: Sort documents by score.
	sort.Slice(docs, func(i, j int) bool {
		return scores[docs[i].Id] > scores[docs[j].Id]
	})

	return docs
}

func calculateIDF(docFreqs map[string]int, docsNum int) map[string]float64 {
	N := float64(docsNum)
	idf := make(map[string]float64)
	for term, freq := range docFreqs {
		n_q := float64(freq)
		idf[term] = math.Log((N - n_q + 0.5) / (n_q + 0.5) + 1)
	}
	
	return idf
}