package core

import (
	"container/heap"
	"fmt"
)

type DocumentID = string

type Document interface {
	ID() DocumentID
	Text() string
}

type BaseDocument struct {
	Name    DocumentID
	Content string
}

func NewBaseDocument(id DocumentID, text string) BaseDocument {
	return BaseDocument{Name: id, Content: text}
}

func (d BaseDocument) ID() string {
	return d.Name
}

func (d BaseDocument) Text() string {
	return d.Content
}

type DocumentScore struct {
	Doc   Document
	Score float64
}

func (ds DocumentScore) String() string {
	return fmt.Sprintf("Score: %f - Doc: %s;", ds.Score, ds.Doc.ID())
}

// TermFrequency stores relative or absolute term frequencies
type TermFrequency map[string]float64

// Top terms, just a test using the heap data structure
func (tf TermFrequency) Top(n int) []string {
	if n > len(tf) {
		n = len(tf)
	}
	tfHeap := &FloatHeap{}
	heap.Init(tfHeap)
	for k, v := range tf {
		heap.Push(tfHeap, FloatHeapItem{Key: k, Value: v})
	}
	terms := make([]string, 0, n)
	for i := 0; i < n; i++ {
		terms = append(terms, heap.Pop(tfHeap).(FloatHeapItem).Key.(string))
	}
	return terms
}

func NewTermFrequency(terms []string) TermFrequency {
	frequencies := make(TermFrequency)
	for _, term := range terms {
		frequencies[term]++
	}
	return frequencies
}

type DocumentData struct {
	Doc         Document
	TermsCount  float64
	Frequencies TermFrequency
}
