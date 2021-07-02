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
	out := "../../test/testdata/index.out"
	analyzer := analysis.SimpleAnalyzer
	idx, _ := LoadDocuments("../../test/testdata/books", analyzer)
	Export(idx, out)
	idx2, _ := Import(out)
	if idx2.analyzer != analyzer {
		t.Error("Imported index should have the same analyzer")
	}
	if idx2.Length() != idx.Length() {
		t.Error("Imported index should have the same length")
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
