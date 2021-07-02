package analysis

import (
	"testing"

	"github.com/ruial/busca/test"
)

func TestSimpleAnalyzer(t *testing.T) {
	text := "The 2 QUICK   Brown-Foxes jumped over the lazy dog's bone..."
	expected := []string{"the", "2", "quick", "brown", "foxes", "jumped", "over", "the", "lazy", "dog's", "bone"}
	result := SimpleAnalyzer.Analyze(text)
	if len(result) != len(expected) {
		t.Errorf("Length: %d - %d, Analyzer results is not expected:\n%s\n%s\n", len(result), len(expected), result, expected)
	}
	if !test.StringArrayEquals(result, expected, true) {
		t.Error("Analyzer result not equal to expected:", result)
	}
}

func TestWhitespaceAnalyzer(t *testing.T) {
	text := "The 2 QUICK   Brown-Foxes jumped over the lazy dog's bone..."
	expected := []string{"The", "2", "QUICK", "Brown-Foxes", "jumped", "over", "the", "lazy", "dog's", "bone..."}
	result := WhiteSpaceAnalyzer.Analyze(text)
	if len(result) != len(expected) {
		t.Errorf("Length: %d - %d, Analyzer results is not expected:\n%s\n%s\n", len(result), len(expected), result, expected)
	}
	if !test.StringArrayEquals(result, expected, true) {
		t.Error("Analyzer result not equal to expected:", result)
	}
}
