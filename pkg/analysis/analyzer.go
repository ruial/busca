package analysis

import (
	"errors"
	"strings"
	"unicode"

	porter "github.com/blevesearch/go-porterstemmer"
	"github.com/blevesearch/segment"
)

var ErrUnavailableStemmer = errors.New("Unavailable stemmer")

type Analyzer interface {
	Analyze(text string) (terms []string)
	GetStopwords() []string
	GetSynonyms() map[string]string
	GetStemmer() string
	String() string
}

type Settings struct {
	Stopwords map[string]struct{}
	Synonyms  map[string]string
	Stemmer   string
}

func NewSettings(stopWords map[string]struct{}, synonyms map[string]string, stemmer string) (s Settings, err error) {
	if stemmer != "" {
		stemmer = strings.ToLower(stemmer)
		if stemmer != "english" {
			err = ErrUnavailableStemmer
			return
		}
	}
	s.Stopwords = make(map[string]struct{})
	for stopword := range stopWords {
		s.Stopwords[s.stem(stopword)] = struct{}{}
	}
	s.Synonyms = make(map[string]string)
	for word, synonym := range synonyms {
		s.Synonyms[s.stem(word)] = s.stem(synonym)
	}
	s.Stemmer = stemmer
	return
}

func (b Settings) stem(word string) string {
	if b.Stemmer == "" {
		return word
	}
	// another very fast stemmer is https://github.com/shabbyrobe/go-porter2
	return string(porter.StemWithoutLowerCasing([]rune(word)))
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

func (b Settings) GetStemmer() string {
	return b.Stemmer
}

func (b Settings) filterTokens(tokens []string) []string {
	result := make([]string, 0, len(tokens))
	for _, token := range tokens {
		token := b.stem(token)
		if synonym, ok := b.Synonyms[token]; ok {
			token = synonym
		}
		if _, ok := b.Stopwords[token]; ok {
			continue
		}
		result = append(result, token)
	}
	return result
}

type StandardAnalyzer struct {
	Settings
}

func (s StandardAnalyzer) Analyze(text string) (result []string) {
	// performs Unicode Text Segmentation as described in Unicode Standard Annex #29
	segmenter := segment.NewSegmenterDirect([]byte(strings.ToLower(text)))
	for segmenter.Segment() {
		if segmenter.Type() > 0 {
			result = append(result, string(segmenter.Bytes()))
		}
	}
	// if err := segmenter.Err(); err != nil {
	// 	log.Println("Segmenter erorr:", err)
	// }
	return s.filterTokens(result)
}

func (s StandardAnalyzer) String() string {
	return "StandardAnalyzer"
}

type SimpleAnalyzer struct {
	Settings
}

func (s SimpleAnalyzer) Analyze(text string) []string {
	// faster than regex.Split or StandardAnalyzer
	// analyzer steps - character filters > tokenizer > token filters
	tokens := strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return !(unicode.IsLetter(r) || unicode.IsNumber(r) || r == rune('\'') || r == rune('’'))
	})
	return s.filterTokens(tokens)
}

func (s SimpleAnalyzer) String() string {
	return "SimpleAnalyzer"
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
