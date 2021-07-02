package search

import (
	"testing"

	"github.com/ruial/busca/pkg/core"
	"github.com/ruial/busca/test"
)

func TestOrFilter(t *testing.T) {
	docTerms := map[string][]core.DocumentID{
		"another": []string{"2", "3"},
		"cool":    []string{"2"},
		"example": []string{"1", "2", "3"},
		"the":     []string{"1"},
		"last":    []string{"4"},
	}

	for _, c := range []struct {
		searchTerms, docIds []string
	}{
		{[]string{"another", "example"}, []string{"1", "2", "3"}},
		{[]string{"the", "last"}, []string{"1", "4"}},
		{[]string{"very", "last"}, []string{"4"}},
		{[]string{"none"}, []string{}},
	} {
		resultIds := OrFilter(c.searchTerms, docTerms)
		if !test.StringArrayEquals(resultIds, c.docIds, false) {
			t.Errorf("Bad results for %s: %s\n", c.searchTerms, resultIds)
		}
	}
}

func TestAndFilter(t *testing.T) {
	docTerms := map[string][]core.DocumentID{
		"another": []string{"2", "3"},
		"cool":    []string{"2"},
		"example": []string{"1", "2", "3"},
		"the":     []string{"1"},
		"last":    []string{"4"},
	}

	for _, c := range []struct {
		searchTerms, docIds []string
	}{
		{[]string{"another", "example"}, []string{"2", "3"}},
		{[]string{"another", "cool", "example"}, []string{"2"}},
		{[]string{"the", "last"}, []string{}},
	} {
		resultIds := AndFilter(c.searchTerms, docTerms)
		if !test.StringArrayEquals(resultIds, c.docIds, false) {
			t.Errorf("Bad results for %s: %s\n", c.searchTerms, resultIds)
		}
	}
}
