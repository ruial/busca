package analysis

import (
	"testing"

	"github.com/ruial/busca/internal/util"
)

func TestStandardAnalyzer(t *testing.T) {
	text := "Alice’s The 2 QUICK   Brown-Foxes jumped over the lazy dog's bone..."
	expected := []string{"alice’s", "the", "2", "quick", "brown", "foxes", "jumped", "over", "the", "lazy", "dog's", "bone"}
	result := StandardAnalyzer.Analyze(text)
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
	result := WhitespaceAnalyzer.Analyze(text)
	if len(result) != len(expected) {
		t.Errorf("Length: %d - %d, Analyzer results is not expected:\n%s\n%s\n", len(result), len(expected), result, expected)
	}
	if !util.StringArrayEquals(result, expected, true) {
		t.Error("Analyzer result not equal to expected:", result)
	}
}
