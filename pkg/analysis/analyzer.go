package analysis

import (
	"strings"
	"unicode"
)

var (
	SimpleAnalyzer     = simpleAnalyzer{}
	WhiteSpaceAnalyzer = whitespaceAnalyzer{}
)

type Analyzer interface {
	Analyze(text string) (terms []string)
	String() string
}

type simpleAnalyzer struct{}

func (s simpleAnalyzer) Analyze(text string) []string {
	// faster than regex.Split
	return strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return !(unicode.IsLetter(r) || unicode.IsNumber(r) || r == rune('\''))
	})
}

func (s simpleAnalyzer) String() string {
	return "SimpleAnalyzer"
}

type whitespaceAnalyzer struct{}

func (w whitespaceAnalyzer) Analyze(text string) []string {
	return strings.Fields(text)
}

func (w whitespaceAnalyzer) String() string {
	return "WhiteSpaceAnalyzer"
}
