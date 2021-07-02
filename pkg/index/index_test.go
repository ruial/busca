package index

import (
	"fmt"
	"io/ioutil"
	"sync"
	"testing"

	"github.com/ruial/busca/pkg/analysis"
	"github.com/ruial/busca/pkg/core"
	"github.com/ruial/busca/pkg/search"
)

func TestAdd(t *testing.T) {
	idx := New(analysis.SimpleAnalyzer)
	doc := core.NewBaseDocument("id", "text")
	if err := idx.AddDocument(doc); err != nil {
		t.Error("First document should not throw error")
	}
	if err := idx.AddDocument(doc); err != ErrDuplicateDocument {
		t.Error("Duplicate should throw error")
	}
	if length := idx.Length(); length != 1 {
		t.Error("Length should be equal to 1")
	}
	if doc != idx.GetDocument(doc.ID()) {
		t.Error("Document should be equal")
	}
	if !idx.HasDocument(doc.ID()) {
		t.Error("Should have document")
	}
}

func TestUpdate(t *testing.T) {
	idx := New(analysis.SimpleAnalyzer)
	doc := core.NewBaseDocument("id", "text")
	if err := idx.UpdateDocument(doc); err != ErrNonExistentDocument {
		t.Error("Update non existing throw error")
	}
	idx.AddDocument(doc)
	doc2 := core.NewBaseDocument("id", "text2")
	if err := idx.UpdateDocument(doc2); err != nil {
		t.Error("Should update existing without error", err)
	}
	if length := idx.Length(); length != 1 {
		t.Error("Length should be equal to 1")
	}
	if doc2 != idx.GetDocument(doc.ID()) {
		t.Error("Document should be equal")
	}
}

func TestUpsert(t *testing.T) {
	idx := New(analysis.SimpleAnalyzer)
	doc := core.NewBaseDocument("id", "text")
	if err := idx.UpsertDocument(doc); err != nil {
		t.Error("Upsert should not throw error on create")
	}
	doc2 := core.NewBaseDocument("id", "text2")
	if err := idx.UpsertDocument(doc2); err != nil {
		t.Error("Should update existing without error", err)
	}
	if length := idx.Length(); length != 1 {
		t.Error("Length should be equal to 1")
	}
	if doc2 != idx.GetDocument(doc.ID()) {
		t.Error("Document should be equal")
	}
}

func TestDelete(t *testing.T) {
	idx := New(analysis.SimpleAnalyzer)
	doc := core.NewBaseDocument("id", "text")
	idx.AddDocument(doc)
	err := idx.DeleteDocument(doc.ID())
	if err != nil {
		t.Error("Should not throw error on deleting valid document")
	}
	err = idx.DeleteDocument(doc.ID())
	if err != ErrNonExistentDocument {
		t.Error("Should throw error on deleting non existent document")
	}
	if idx.HasDocument(doc.ID()) || idx.GetDocument(doc.ID()) != nil {
		t.Error("Should not have document anymore")
	}
	if idx.Length() != 0 || len(idx.terms) != 0 || len(idx.documents) != 0 {
		t.Error("Length should be empty")
	}
}

func TestAddDeleteConcurrent(t *testing.T) {
	idx := New(analysis.SimpleAnalyzer)
	// repeat some document ids to simulate collisions
	length := 100
	threads := 1000
	var wg sync.WaitGroup
	for i := 0; i <= threads; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			doc := core.NewBaseDocument(fmt.Sprint(i%length), "some text")
			idx.AddDocument(doc)
		}(i)
	}
	wg.Wait()
	idxLength := idx.Length()
	if idxLength != length {
		t.Errorf("Expected length was %d, received %d", idxLength, length)
	}

	for i := 0; i <= length; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			idx.DeleteDocument(fmt.Sprint(i % length))
		}(i)
	}
	wg.Wait()
	idxLength = idx.Length()
	if idxLength != 0 {
		t.Error("Expected length to be zero")
	}
}

func TestSearch(t *testing.T) {
	idx := New(analysis.SimpleAnalyzer)
	for _, doc := range []core.BaseDocument{
		core.NewBaseDocument("1", "the first example example"),
		core.NewBaseDocument("2", "another cool example"),
		core.NewBaseDocument("3", "another example"),
		core.NewBaseDocument("4", "last"),
	} {
		idx.AddDocument(doc)
	}
	// normal case
	docs := idx.SearchDocuments("cool example", search.AndFilter, search.TfIdfRanker(search.TfWeightDefault, search.IdfWeightDefault))
	if !(len(docs) == 1) && docs[0].Doc.ID() == "2" {
		t.Error("Invalid search results:", docs)
	}
	// without filter should return all documents
	docs = idx.SearchDocuments("cool example", nil, search.TfIdfRanker(search.TfWeightDefault, search.IdfWeightDefault))
	if len(docs) != 4 {
		t.Error("Invalid search results:", docs)
	}
	// with MinMatchFilter at 0 also should return all documents
	docs = idx.SearchDocuments("cool example", search.MinMatchFilter(0), nil)
	if len(docs) != 4 {
		t.Error("Invalid search results:", docs)
	}
	// without ranking all results have same score
	docs = idx.SearchDocuments("cool example", search.OrFilter, nil)
	if len(docs) != 3 {
		t.Error("Invalid search results: ", docs)
	}
	for _, doc := range docs {
		if doc.Score != 0 {
			t.Error("Invalid doc:", doc)
		}
	}
}

func BenchmarkAdd(b *testing.B) {
	idx := New(analysis.SimpleAnalyzer)
	bytes, err := ioutil.ReadFile("../../test/testdata/books/Dracula.txt")
	if err != nil {
		b.Error("File not found")
	}
	// Repeat ids to simulate some collisions and trim the text
	length := 1000
	text := string(bytes[:10000])
	b.Logf("Adding %d documents", b.N)
	var wg sync.WaitGroup
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			doc := core.NewBaseDocument(fmt.Sprint(i%length), text)
			idx.AddDocument(doc)
		}(i)
	}
	wg.Wait()
	if b.N < length {
		length = b.N
	}
	idxLength := idx.Length()
	if idxLength != length {
		b.Errorf("Expected length was %d, received %d", length, b.N)
	}
}
