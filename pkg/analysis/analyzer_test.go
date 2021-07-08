package analysis

import (
	"testing"

	"github.com/ruial/busca/internal/util"
)

func TestStandardAnalyzer(t *testing.T) {
	text := "Alice’s The 2 QUICK and rapid   Brown-Foxes 'jumped' over the lazy dog's bone..."
	expected := []string{"alice’s", "2", "fast", "fast", "brown", "foxes", "jumped", "over", "lazy", "dog's", "bone"}
	stopwords := map[string]struct{}{
		"the": {},
		"and": {},
	}
	synonyms := map[string]string{
		"quick": "fast",
		"rapid": "fast",
	}
	settings, _ := NewSettings(stopwords, synonyms, "")
	analyzer := StandardAnalyzer{Settings: settings}
	if len(analyzer.GetStopwords()) != len(stopwords) {
		t.Error("Expected analyzer stopwords array length to be equal to set")
	}
	result := analyzer.Analyze(text)
	if !util.StringArrayEquals(result, expected, true) {
		t.Error("Analyzer result not equal to expected:", result)
	}
}

func TestStandardAnalyzerStem(t *testing.T) {
	text := "jumped over the lazy cat"
	expected := []string{"leap", "over", "lazi", "cat"}
	stopwords := map[string]struct{}{"the": {}}
	synonyms := map[string]string{"jump": "leap"}
	settings, _ := NewSettings(stopwords, synonyms, "English")
	analyzer := StandardAnalyzer{Settings: settings}
	result := analyzer.Analyze(text)
	if !util.StringArrayEquals(result, expected, true) {
		t.Error("Analyzer result not equal to expected:", result)
	}
}

func TestSimpleAnalyzer(t *testing.T) {
	text := "Alice’s The 2 QUICK   Brown-Foxes 'jumped' over the lazy dog's bone..."
	expected := []string{"alice’s", "the", "2", "quick", "brown", "foxes", "'jumped'", "over", "the", "lazy", "dog's", "bone"}
	result := SimpleAnalyzer{}.Analyze(text)
	if !util.StringArrayEquals(result, expected, true) {
		t.Error("Analyzer result not equal to expected:", result)
	}
}

func TestWhitespaceAnalyzer(t *testing.T) {
	text := "Alice’s The 2 QUICK   Brown-Foxes 'jumped' over the lazy dog's bone..."
	expected := []string{"Alice’s", "The", "2", "QUICK", "Brown-Foxes", "'jumped'", "over", "the", "lazy", "dog's", "bone..."}
	result := WhitespaceAnalyzer{}.Analyze(text)
	if !util.StringArrayEquals(result, expected, true) {
		t.Error("Analyzer result not equal to expected:", result)
	}
}
