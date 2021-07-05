package analysis

import (
	"strings"
	"unicode"
)

type Analyzer interface {
	Analyze(text string) (terms []string)
	GetStopwords() []string
	GetSynonyms() map[string]string
	String() string
}

type Settings struct {
	Stopwords map[string]struct{}
	Synonyms  map[string]string
}

func (b Settings) GetStopwords() []string {
	list := make([]string, 0, len(b.Stopwords))
	for stopword := range b.Stopwords {
		list = append(list, stopword)
	}
	return list
}

func (b Settings) GetSynonyms() map[string]string {
	return b.Synonyms
}

func (b Settings) filterTokens(tokens []string) (result []string) {
	for _, token := range tokens {
		if synonym, ok := b.Synonyms[token]; ok {
			token = synonym
		}
		if _, ok := b.Stopwords[token]; ok {
			continue
		}
		result = append(result, token)
	}
	return
}

type StandardAnalyzer struct {
	Settings
}

func (s StandardAnalyzer) Analyze(text string) (result []string) {
	// faster than regex.Split
	// analyzer steps - character filters > tokenizer > token filters
	tokens := strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return !(unicode.IsLetter(r) || unicode.IsNumber(r) || r == rune('\'') || r == rune('’'))
	})
	return s.filterTokens(tokens)
}

func (s StandardAnalyzer) String() string {
	return "StandardAnalyzer"
}

type WhitespaceAnalyzer struct {
	Settings
}

func (w WhitespaceAnalyzer) Analyze(text string) []string {
	return w.filterTokens(strings.Fields(text))
}

func (w WhitespaceAnalyzer) String() string {
	return "WhitespaceAnalyzer"
}
