package preprocessing

import (
	"strings"

	"github.com/bbalet/stopwords"
	"github.com/kljensen/snowball"
	"github.com/aaaton/golem/v4"
	"github.com/aaaton/golem/v4/dicts/en"
)

func punctuationRemoval(document string) string {
	punctuation := `!()-[]{};:'"\,<>./?@#$%^&*_~`
	var result strings.Builder

	for _, char := range document {
		if !strings.ContainsRune(punctuation, char) {
			result.WriteRune(char)
		}
	}

	return result.String()
}

func stopWordRemoval(document string) string {
	return stopwords.CleanString(document, "en", true)
}

func caseFold(document string) string {
	return strings.ToLower(document)
}

func lemmatize(document string) string {
	lemmatizer, err := golem.New(en.New())
	if err != nil {
		panic(err)
	}

	return lemmatizer.Lemma(document)
}

func stem(document string) string {
	words := strings.Fields(document)
	var result []string
	for _, word := range words {
		stem, err := snowball.Stem(word, "english", true)
		if err != nil {
			stem = word
		}
		result = append(result, stem)
	}

	return strings.Join(result, " ")
}

func Preprocess(document string) []string {
	document = punctuationRemoval(document)
	document = stopWordRemoval(document)
	document = caseFold(document)
	document = lemmatize(document)
	document = stem(document)

	return strings.Split(document, "")
}
