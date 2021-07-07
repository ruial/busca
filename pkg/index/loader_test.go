package index

import (
	"os"
	"testing"

	"github.com/ruial/busca/pkg/analysis"
	"github.com/ruial/busca/pkg/search"
)

func TestExportImport(t *testing.T) {
	// To run only unit tests: go test ./... -v -short
	// another alternative for integration tests would be to check an environment variable or use build tags
	if testing.Short() {
		t.Skip("Skipping integration test")
	}
	out := "../../testdata/index.out"
	analyzer := analysis.SimpleAnalyzer{Settings: analysis.Settings{Stopwords: map[string]struct{}{"the": {}}}}
	opts := Opts{Analyzer: analyzer, FuzzyDepth: 1, FuzzyMinOccurrences: 50}
	idx, _ := LoadDocuments("../../testdata/books", opts)
	Export(idx, out)
	idx2, _ := Import(out)
	// could do full equality check, == not enough as struct has map/slice, would have to add an Equal method
	if len(idx2.analyzer.(analysis.SimpleAnalyzer).Stopwords) != len(analyzer.Stopwords) {
		t.Error("Imported index should have the same analyzer")
	}
	if idx2.Length() != idx.Length() {
		t.Error("Imported index should have the same length")
	}
	if len(idx.GetSpellSuggestions([]string{"mor"}, 3)) != len(idx2.GetSpellSuggestions([]string{"mor"}, 3)) {
		t.Error("Expected fuzzy suggestions to be the same")
	}
	tfidfRanker := search.TfIdfRanker(search.TfWeightDefault, search.IdfWeightDefault)
	expected := idx.SearchDocuments("crime detective", nil, tfidfRanker)
	result := idx2.SearchDocuments("crime detective", nil, tfidfRanker)
	// only compare first 3 results because last ones have 0 score and order is not guaranteed
	for i := 0; i < 3; i++ {
		if !(expected[i].Doc.ID() == result[i].Doc.ID() && expected[i].Score == result[i].Score) {
			t.Error("Imported index search results are not the same")
		}
	}
	if err := os.Remove(out); err != nil {
		t.Error("Unable to remove index file")
	}
}
