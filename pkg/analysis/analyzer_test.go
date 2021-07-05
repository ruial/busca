package analysis

import (
	"testing"

	"github.com/ruial/busca/internal/util"
)

func TestStandardAnalyzer(t *testing.T) {
	text := "Alice’s The 2 QUICK and rapid   Brown-Foxes jumped over the lazy dog's bone..."
	expected := []string{"alice’s", "2", "fast", "fast", "brown", "foxes", "jumped", "over", "lazy", "dog's", "bone"}
	stopwords := map[string]struct{}{
		"the": {},
		"and": {},
	}
	synonyms := map[string]string{
		"quick": "fast",
		"rapid": "fast",
	}
	analyzer := StandardAnalyzer{Settings{Stopwords: stopwords, Synonyms: synonyms}}
	if len(analyzer.GetStopwords()) != len(stopwords) {
		t.Error("Expected analyzer stopwords array length to be equal to set")
	}
	result := analyzer.Analyze(text)
	if len(result) != len(expected) {
		t.Errorf("Length: %d - %d, Analyzer results is not expected:\n%s\n%s\n", len(result), len(expected), result, expected)
	}
	if !util.StringArrayEquals(result, expected, true) {
		t.Error("Analyzer result not equal to expected:", result)
	}
}

func TestWhitespaceAnalyzer(t *testing.T) {
	text := "Alice’s The 2 QUICK   Brown-Foxes jumped over the lazy dog's bone..."
	expected := []string{"Alice’s", "The", "2", "QUICK", "Brown-Foxes", "jumped", "over", "the", "lazy", "dog's", "bone..."}
	result := WhitespaceAnalyzer{}.Analyze(text)
	if len(result) != len(expected) {
		t.Errorf("Length: %d - %d, Analyzer results is not expected:\n%s\n%s\n", len(result), len(expected), result, expected)
	}
	if !util.StringArrayEquals(result, expected, true) {
		t.Error("Analyzer result not equal to expected:", result)
	}
}
