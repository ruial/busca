package analysis

import (
	"strings"
	"unicode"
)

var (
	StandardAnalyzer   = standardAnalyzer{}
	WhiteSpaceAnalyzer = whitespaceAnalyzer{}
)

type Analyzer interface {
	Analyze(text string) (terms []string)
	String() string
}

type standardAnalyzer struct{}

func (s standardAnalyzer) Analyze(text string) []string {
	// faster than regex.Split
	// analyzer steps - character filters > tokenizer > token filters
	return strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return !(unicode.IsLetter(r) || unicode.IsNumber(r) || r == rune('\'') || r == rune('’'))
	})
}

func (s standardAnalyzer) String() string {
	return "StandardAnalyzer"
}

type whitespaceAnalyzer struct{}

func (w whitespaceAnalyzer) Analyze(text string) []string {
	return strings.Fields(text)
}

func (w whitespaceAnalyzer) String() string {
	return "WhiteSpaceAnalyzer"
}
