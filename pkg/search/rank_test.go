package search

import (
	"math"
	"testing"

	"github.com/ruial/busca/pkg/core"
)

const floatEqualityThreshold = 1e-6

func floatEqual(a, b float64) bool {
	return math.Abs(a-b) < floatEqualityThreshold
}

func TestRank(t *testing.T) {
	docsData := map[core.DocumentID]core.DocumentData{
		"1": {Doc: core.NewBaseDocument("1", ""), TermsCount: 4, Frequencies: core.NewTermFrequency([]string{"the", "first", "example", "example"})},
		"2": {Doc: core.NewBaseDocument("2", ""), TermsCount: 3, Frequencies: core.NewTermFrequency([]string{"another", "cool", "example"})},
		"3": {Doc: core.NewBaseDocument("3", ""), TermsCount: 2, Frequencies: core.NewTermFrequency([]string{"another", "example"})},
		"4": {Doc: core.NewBaseDocument("4", ""), TermsCount: 1, Frequencies: core.NewTermFrequency([]string{"last"})},
	}
	// test relative term frequency and default idf formula
	ranker := TfIdfRanker(TfWeightDefault, IdfWeightDefault)
	scores := ranker([]string{"great", "first", "example"}, docsData)
	expected := []core.DocumentScore{
		{Doc: core.NewBaseDocument("1", ""), Score: 0.212984},
		{Doc: core.NewBaseDocument("3", ""), Score: 0.062469},
		{Doc: core.NewBaseDocument("2", ""), Score: 0.041646},
		{Doc: core.NewBaseDocument("4", ""), Score: 0},
	}
	for i := range scores {
		if !(scores[i].Doc.ID() == expected[i].Doc.ID() && floatEqual(scores[i].Score, expected[i].Score)) {
			t.Errorf("Invalid ranking: %s %s", scores[i], expected[i])
		}
	}
	// test log normalized term frequency and smooth idf formula
	ranker = TfIdfRanker(TfWeightLog, IdfWeightSmooth)
	scores = ranker([]string{"great", "first", "example"}, docsData)
	expected = []core.DocumentScore{
		{Doc: core.NewBaseDocument("1", ""), Score: 0.944181},
		{Doc: core.NewBaseDocument("3", ""), Score: 0.330203},
		{Doc: core.NewBaseDocument("2", ""), Score: 0.330203},
		{Doc: core.NewBaseDocument("4", ""), Score: 0},
	}
	for i := range scores {
		if !floatEqual(scores[i].Score, expected[i].Score) {
			t.Errorf("Invalid ranking: %s %s", scores[i], expected[i])
		}
	}
}
