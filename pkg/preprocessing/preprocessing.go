package preprocessing

import (
	"bufio"
	"os"
	"strings"
	"sync"

	"github.com/kljensen/snowball"
	"github.com/aaaton/golem/v4"
	"github.com/aaaton/golem/v4/dicts/en"
	
	// Use Prose v2 for advanced NLP processing
	"github.com/jdkato/prose/v2"
)

// Global variables for caching stop words
var (
	stopWordsSet map[string]bool
	stopWordsOnce sync.Once
	stopWordsError error
)

// ===== Prose v2 Advanced Tokenizer =====

type ProseTokenizer struct {
	preserveCase bool
	filterPOS    bool
}

func NewProseTokenizer(preserveCase, filterPOS bool) *ProseTokenizer {
	return &ProseTokenizer{
		preserveCase: preserveCase,
		filterPOS:    filterPOS,
	}
}

func (pt *ProseTokenizer) Tokenize(text string) ([]string, error) {
	doc, err := prose.NewDocument(text)
	if err != nil {
		return nil, err
	}

	var tokens []string
	
	if pt.filterPOS {
		// Use POS tagging for intelligent filtering
		for _, token := range doc.Tokens() {
			if pt.shouldKeepToken(token.Text, token.Tag) {
				tokenText := token.Text
				if !pt.preserveCase {
					tokenText = strings.ToLower(tokenText)
				}
				tokens = append(tokens, tokenText)
			}
		}
	} else {
		// Simple tokenization
		for _, token := range doc.Tokens() {
			tokenText := strings.TrimSpace(token.Text)
			if tokenText != "" {
				if !pt.preserveCase {
					tokenText = strings.ToLower(tokenText)
				}
				tokens = append(tokens, tokenText)
			}
		}
	}

	return tokens, nil
}

func (pt *ProseTokenizer) TokenizeWithPOS(text string) ([]string, []string, error) {
	doc, err := prose.NewDocument(text)
	if err != nil {
		return nil, nil, err
	}

	var tokens, tags []string
	for _, token := range doc.Tokens() {
		tokenText := strings.TrimSpace(token.Text)
		if tokenText != "" {
			if !pt.preserveCase {
				tokenText = strings.ToLower(tokenText)
			}
			tokens = append(tokens, tokenText)
			tags = append(tags, token.Tag)
		}
	}

	return tokens, tags, nil
}

func (pt *ProseTokenizer) TokenizeWithEntities(text string) ([]string, []prose.Entity, error) {
	doc, err := prose.NewDocument(text)
	if err != nil {
		return nil, nil, err
	}

	tokens, _ := pt.Tokenize(text)
	entities := doc.Entities()
	
	return tokens, entities, nil
}

func (pt *ProseTokenizer) shouldKeepToken(text, pos string) bool {
	text = strings.TrimSpace(strings.ToLower(text))
	if text == "" {
		return false
	}

	// Keep numbers and model numbers
	if isNumericValue(text) || isModelNumber(text) {
		return true
	}

	// Keep important product words
	if isImportantProductWord(text) {
		return true
	}

	// Filter based on POS tags
	switch pos {
	case "NN", "NNS", "NNP", "NNPS": // Nouns
		return true
	case "JJ", "JJR", "JJS": // Adjectives
		return true
	case "VB", "VBG", "VBN", "VBD", "VBZ": // Verbs
		return true
	case "CD": // Numbers
		return true
	case "FW": // Foreign words
		return true
	case "RB", "RBR", "RBS": // Adverbs (for words like "very", "extremely")
		return len(text) > 3 // Keep longer adverbs
	default:
		return false
	}
}

// ===== Configuration for different use cases =====

type TokenizerConfig struct {
	PreserveCase bool
	FilterPOS    bool
	MaxLength    int
	MinLength    int
}

func tokenizeWithProse(document string, config TokenizerConfig) []string {
	tokenizer := NewProseTokenizer(config.PreserveCase, config.FilterPOS)
	
	tokens, err := tokenizer.Tokenize(document)
	if err != nil {
		// Fallback to basic tokenization
		return fallbackTokenize(document)
	}
	
	// Apply length filtering
	var filtered []string
	for _, token := range tokens {
		if len(token) >= config.MinLength && len(token) <= config.MaxLength {
			filtered = append(filtered, token)
		}
	}
	
	return filtered
}

// ===== Helper functions =====

func punctuationRemoval(document string) string {
	punctuation := `!()[]{};:,'"\<>/?#$^&*_~`
	var result strings.Builder

	for i, char := range document {
		if !strings.ContainsRune(punctuation, char) {
			result.WriteRune(char)
		} else if char == '.' {
			runes := []rune(document)
			if i > 0 && i < len(runes)-1 && 
			   isDigitRune(runes[i-1]) && isDigitRune(runes[i+1]) {
				result.WriteRune(char)
			} else {
				result.WriteRune(' ')
			}
		} else if char == '-' {
			result.WriteRune(' ')
		} else {
			result.WriteRune(' ')
		}
	}

	return result.String()
}

func isDigitRune(r rune) bool {
	return r >= '0' && r <= '9'
}

func loadStopWords(filename string) (map[string]bool, error) {
	stopWordsOnce.Do(func() {
		stopWordsSet = make(map[string]bool)
		
		file, err := os.Open(filename)
		if err != nil {
			stopWordsError = err
			return
		}
		defer file.Close()
		
		scanner := bufio.NewScanner(file)
		buf := make([]byte, 0, 64*1024)
		scanner.Buffer(buf, 1024*1024)
		
		for scanner.Scan() {
			word := strings.TrimSpace(scanner.Text())
			if word != "" && !strings.HasPrefix(word, "#") {
				stopWordsSet[strings.ToLower(word)] = true
			}
		}
		
		if err := scanner.Err(); err != nil {
			stopWordsError = err
		}
	})
	
	return stopWordsSet, stopWordsError
}

func stopWordRemoval(document string) string {
	stopWords, err := loadStopWords("stopwords.txt")
	if err != nil {
		return stopWordRemovalFallback(document)
	}
	
	words := strings.Fields(document)
	if len(words) == 0 {
		return document
	}
	
	result := make([]string, 0, len(words))
	
	for _, word := range words {
		wordTrimmed := strings.TrimSpace(word)
		if wordTrimmed == "" {
			continue
		}
		
		lowerWord := strings.ToLower(wordTrimmed)
		
		if isNumericValue(wordTrimmed) || isModelNumber(wordTrimmed) || 
		   isImportantProductWord(lowerWord) || !stopWords[lowerWord] {
			result = append(result, wordTrimmed)
		}
	}
	
	var builder strings.Builder
	builder.Grow(len(document))
	
	for i, word := range result {
		if i > 0 {
			builder.WriteByte(' ')
		}
		builder.WriteString(word)
	}
	
	return builder.String()
}

func stopWordRemovalFallback(document string) string {
	basicStopWords := map[string]bool{
		"a": true, "an": true, "and": true, "are": true, "as": true, "at": true,
		"be": true, "been": true, "by": true, "for": true, "from": true,
		"has": true, "he": true, "in": true, "is": true, "it": true,
		"its": true, "of": true, "on": true, "that": true, "the": true,
		"to": true, "was": true, "will": true, "with": true, "have": true,
		"his": true, "her": true, "him": true, "she": true, "they": true,
		"we": true, "you": true, "i": true, "me": true, "my": true,
		"this": true, "these": true, "those": true, "than": true,
	}
	
	words := strings.Fields(document)
	result := make([]string, 0, len(words))
	
	for _, word := range words {
		lowerWord := strings.ToLower(strings.TrimSpace(word))
		if !basicStopWords[lowerWord] && len(lowerWord) > 0 {
			result = append(result, word)
		}
	}
	
	return strings.Join(result, " ")
}

func caseFold(document string) string {
	return strings.ToLower(document)
}

func lemmatize(document string) string {
	lemmatizer, err := golem.New(en.New())
	if err != nil {
		return document // Return original document instead of panic
	}
	return lemmatizer.Lemma(document)
}

func stem(document string) string {
	words := strings.Fields(document)
	var result []string
	
	noStemWords := map[string]bool{
		"inverter": true, "convertible": true, "filter": true,
		"protection": true, "cooling": true, "model": true,
		"copper": true, "white": true, "anti": true, "virus": true,
		"viral": true, "dual": true, "super": true, "flexicool": true,
		"conditioner": true, "portable": true, "heating": true,
		"saving": true, "energy": true,
	}
	
	for _, word := range words {
		lowerWord := strings.ToLower(word)
		
		if isNumericValue(word) || isModelNumber(word) || noStemWords[lowerWord] {
			result = append(result, word)
			continue
		}
		
		stem, err := snowball.Stem(word, "english", true)
		if err != nil {
			stem = word
		}
		result = append(result, stem)
	}

	return strings.Join(result, " ")
}

func isNumericValue(s string) bool {
	if len(s) == 0 {
		return false
	}
	
	hasDigit := false
	hasDot := false
	
	for _, char := range s {
		if char >= '0' && char <= '9' {
			hasDigit = true
		} else if char == '.' {
			if hasDot {
				return false
			}
			hasDot = true
		} else {
			return false
		}
	}
	
	return hasDigit
}

func isModelNumber(s string) bool {
	if len(s) < 4 {
		return false
	}
	
	upperCount := 0
	digitCount := 0
	
	for _, char := range s {
		if char >= 'A' && char <= 'Z' {
			upperCount++
		} else if char >= '0' && char <= '9' {
			digitCount++
		} else if char >= 'a' && char <= 'z' {
			// Allow lowercase letters
		} else {
			return false
		}
	}
	
	return upperCount >= 2 && digitCount >= 2
}

func isImportantProductWord(word string) bool {
	importantWords := map[string]bool{
		"ac": true, "air": true, "conditioner": true, "inverter": true,
		"split": true, "window": true, "portable": true, "copper": true,
		"filter": true, "anti": true, "viral": true, "virus": true,
		"protection": true, "cooling": true, "heating": true, "energy": true,
		"saving": true, "star": true, "ton": true, "dual": true,
		"convertible": true, "smart": true, "wifi": true, "bluetooth": true,
		"white": true, "black": true, "silver": true, "blue": true,
		"model": true, "super": true, "hd": true, "pm": true,
		"flexicool": true, "dxi": true, "ester": true, "ai": true,
		"remote": true, "control": true, "timer": true, "display": true,
		"compressor": true, "refrigerant": true, "btu": true,
	}
	
	return importantWords[word]
}

func fallbackTokenize(document string) []string {
	tokens := strings.Fields(strings.TrimSpace(document))
	var result []string
	
	for _, token := range tokens {
		token = strings.TrimSpace(token)
		if len(token) > 0 {
			result = append(result, token)
		}
	}
	
	return result
}

// ===== Main preprocessing functions =====

// Standard preprocessing with Prose v2
func Preprocess(document string) []string {
	document = punctuationRemoval(document)
	document = stopWordRemoval(document)
	document = caseFold(document)
	document = lemmatize(document)
	document = stem(document)

	// Use Prose v2 tokenizer with intelligent filtering
	config := TokenizerConfig{
		PreserveCase: false,
		FilterPOS:    true,
		MaxLength:    50,
		MinLength:    1,
	}
	
	result := tokenizeWithProse(document, config)
	return result
}

// Optimized preprocessing for large datasets
func PreprocessForLargeDataset(document string) []string {
	document = punctuationRemoval(document)
	document = stopWordRemoval(document)
	document = caseFold(document)
	// Skip lemmatize and stem to improve performance
	
	// Use basic Prose v2 tokenization
	config := TokenizerConfig{
		PreserveCase: false,
		FilterPOS:    false, // Skip POS filtering for speed
		MaxLength:    100,
		MinLength:    2,
	}
	
	result := tokenizeWithProse(document, config)
	return result
}

// Lightweight preprocessing (suitable for resource-constrained environments)
func PreprocessLightweight(document string) []string {
	document = punctuationRemoval(document)
	document = stopWordRemovalFallback(document) // Use built-in stop words
	document = caseFold(document)
	// Skip lemmatize and stem
	
	config := TokenizerConfig{
		PreserveCase: false,
		FilterPOS:    true,
		MaxLength:    30,
		MinLength:    2,
	}
	
	result := tokenizeWithProse(document, config)
	return result
}

// Advanced preprocessing with POS tagging and entity recognition
func PreprocessWithAnalysis(document string) ([]string, []string, []prose.Entity, error) {
	// Basic preprocessing
	document = punctuationRemoval(document)
	document = caseFold(document)
	
	// Use Prose v2 for advanced analysis
	tokenizer := NewProseTokenizer(false, true)
	
	tokens, posTags, err := tokenizer.TokenizeWithPOS(document)
	if err != nil {
		return nil, nil, nil, err
	}
	
	_, entities, err := tokenizer.TokenizeWithEntities(document)
	if err != nil {
		return tokens, posTags, nil, err
	}
	
	return tokens, posTags, entities, nil
}