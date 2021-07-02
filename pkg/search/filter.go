package search

import (
	"math"

	"github.com/ruial/busca/pkg/core"
)

type Filter func(terms []string, termsMap map[string][]core.DocumentID) (docs []core.DocumentID)

var (
	OrFilter  = MinMatchFilter(1)
	AndFilter = MinMatchFilter(math.MaxInt64)
)

func MinMatchFilter(minCount int) Filter {
	if minCount <= 0 {
		return nil
	}
	return func(terms []string, termsMap map[string][]core.DocumentID) (docs []core.DocumentID) {
		docsUnique := make(map[core.DocumentID]int)
		for _, term := range terms {
			for _, doc := range termsMap[term] {
				docsUnique[doc]++
			}
		}
		min := minCount
		if min > len(terms) {
			min = len(terms)
		}
		for doc, matchCount := range docsUnique {
			if matchCount >= min {
				docs = append(docs, doc)
			}
		}
		return
	}
}
