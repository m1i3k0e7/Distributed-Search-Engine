package ranking

import (
	// "log"
	"math"
	"sort"

	search_proto "github.com/m1i3k0e7/distributed-search-engine/api/proto/search"
)

// RankDocumentsByTFIDF ranks documents based on TF-IDF and cosine similarity.
func RankDocumentsByTFIDF(query string, docs []*search_proto.Product) []*search_proto.Product {
	if len(docs) == 0 || query == "" {
		return docs
	}

	// 1. Preprocess query and all documents once
	processedQuery := preprocess(query)
	processedDocs := make(map[string][]string)
	docTermCounts := make(map[string]map[string]int)
	for _, doc := range docs {
		tokens := preprocess(doc.Name)
		processedDocs[doc.Id] = tokens
		
		counts := make(map[string]int)
		for _, token := range tokens {
			counts[token]++
		}
		docTermCounts[doc.Id] = counts
	}

	// 2. Build vocabulary and calculate IDF for each term
	vocab := make(map[string]bool)
	for _, tokens := range processedDocs {
		for _, token := range tokens {
			vocab[token] = true
		}
	}

	// Create a sorted list of vocabulary terms to ensure consistent vector order
	sortedVocab := make([]string, 0, len(vocab))
	for term := range vocab {
		sortedVocab = append(sortedVocab, term)
	}
	sort.Strings(sortedVocab)

	idf := make(map[string]float64)
	totalDocs := float64(len(docs))
	for _, term := range sortedVocab {
		idf[term] = inverseDocumentFrequency(term, totalDocs, docTermCounts)
	}

	// 3. Build TF-IDF vectors for the query and all documents
	queryTermCounts := make(map[string]int)
	for _, token := range processedQuery {
		queryTermCounts[token]++
	}
	queryVector := buildTFIDFVector(len(processedQuery), queryTermCounts, sortedVocab, idf)

	docVectors := make(map[string][]float64)
	for _, doc := range docs {
		docVectors[doc.Id] = buildTFIDFVector(len(processedDocs[doc.Id]), docTermCounts[doc.Id], sortedVocab, idf)
	}

	// 4. Calculate cosine similarity and rank documents
	scores := make(map[string]float64)
	for id, vec := range docVectors {
		scores[id] = cosineSimilarity(queryVector, vec)
	}
	
	sort.SliceStable(docs, func(i, j int) bool {
		return scores[docs[i].Id] > scores[docs[j].Id]
	})

	return docs
}

func inverseDocumentFrequency(term string, totalDocs float64, docTermCounts map[string]map[string]int) float64 {
	docFrequency := 0
	for _, counts := range docTermCounts {
		if _, exists := counts[term]; exists {
			docFrequency++
		}
	}

	return math.Log(totalDocs / (1.0 + float64(docFrequency)))
}

func buildTFIDFVector(tokensCnt int, termCounts map[string]int, sortedVocab []string, idf map[string]float64) []float64 {
	vector := make([]float64, 0, len(sortedVocab))
	if tokensCnt == 0 {
		return vector
	}

	for i, term := range sortedVocab {
		tf := float64(termCounts[term]) / float64(tokensCnt)
		vector[i] = tf * idf[term]
	}

	return vector
}

// CosineSimilarity calculates the cosine similarity between two vectors.
func cosineSimilarity(vecA, vecB []float64) float64 {
	if len(vecA) != len(vecB) {
		return 0.0
	}

	var dotProduct, magnitudeA, magnitudeB float64
	for i := range vecA {
		dotProduct += vecA[i] * vecB[i]
		magnitudeA += vecA[i] * vecA[i]
		magnitudeB += vecB[i] * vecB[i]
	}

	if magnitudeA == 0 || magnitudeB == 0 {
		return 0.0
	}

	return dotProduct / (math.Sqrt(magnitudeA) * math.Sqrt(magnitudeB))
}