package index

import (
	"errors"
	"sync"

	"github.com/ruial/busca/pkg/analysis"
	"github.com/ruial/busca/pkg/core"
	"github.com/ruial/busca/pkg/search"
	"github.com/sajari/fuzzy"
)

var (
	ErrDuplicateDocument   = errors.New("Cannot index duplicate document")
	ErrNonExistentDocument = errors.New("Document does not exist")
)

type indexMode int

const (
	createMode indexMode = iota
	updateMode
	upsertMode
)

type Opts struct {
	Analyzer            analysis.Analyzer
	FuzzyMinOccurrences int
	FuzzyDepth          int
}

type Index struct {
	terms      map[string][]core.DocumentID
	documents  map[core.DocumentID]core.DocumentData
	analyzer   analysis.Analyzer
	fuzzyModel *fuzzy.Model
	docsMutex  sync.RWMutex
}

func New(opts Opts) *Index {
	if opts.FuzzyMinOccurrences < 1 {
		opts.FuzzyMinOccurrences = fuzzy.SpellThresholdDefault
	}
	if opts.FuzzyDepth < 0 {
		opts.FuzzyDepth = fuzzy.SpellDepthDefault
	}
	index := new(Index)
	index.terms = make(map[string][]core.DocumentID)
	index.documents = make(map[core.DocumentID]core.DocumentData)
	index.analyzer = opts.Analyzer
	index.fuzzyModel = fuzzy.NewModel()
	index.fuzzyModel.SetThreshold(opts.FuzzyMinOccurrences)
	index.fuzzyModel.SetDepth(opts.FuzzyDepth)
	index.fuzzyModel.SetUseAutocomplete(false)
	return index
}

func (i *Index) addDocument(document core.Document, idxMode indexMode) error {
	// this part is lock free to improve speed, can be measured with pprof and benchmarks
	// if there are many conflicts, it would be more performant to do the check first, unlock, call the analyzer and lock again
	terms := i.analyzer.Analyze(document.Text())
	frequencies := core.NewTermFrequency(terms)

	// an alternative to locks would be to use channels and get/set the index on a single thread
	// https://stackoverflow.com/questions/27236827/idiomatic-way-to-make-a-request-response-communication-using-channels
	// https://stackoverflow.com/questions/28650702/is-it-thread-safe-to-have-multiple-channels-for-communicating-with-a-single-shar
	i.docsMutex.Lock()
	defer i.docsMutex.Unlock()
	id := document.ID()
	if i.hasDocument(id) {
		if idxMode == createMode {
			return ErrDuplicateDocument
		} else if idxMode == upsertMode {
			i.deleteDocument(id)
		}
	} else if idxMode == updateMode {
		return ErrNonExistentDocument
	}

	var termsCount float64
	for t, f := range frequencies {
		i.terms[t] = append(i.terms[t], id)
		termsCount += f
	}
	i.documents[id] = core.DocumentData{Doc: document, Frequencies: frequencies, TermsCount: termsCount}
	if i.fuzzyModel.Depth > 0 {
		i.fuzzyModel.Train(terms)
	}
	return nil
}

func (i *Index) AddDocument(document core.Document) error {
	return i.addDocument(document, createMode)
}

func (i *Index) UpdateDocument(document core.Document) error {
	return i.addDocument(document, updateMode)
}

func (i *Index) UpsertDocument(document core.Document) error {
	return i.addDocument(document, upsertMode)
}

func (i *Index) deleteDocument(id core.DocumentID) error {
	// private method without locking for use in other methods because Go does not have recursive/reentrant mutex
	if !i.hasDocument(id) {
		return ErrNonExistentDocument
	}
	for term := range i.documents[id].Frequencies {
		newDocs := []string{}
		for _, docId := range i.terms[term] {
			if docId != id {
				newDocs = append(newDocs, docId)
			}
		}
		if len(newDocs) > 0 {
			i.terms[term] = newDocs
		} else {
			delete(i.terms, term)
		}
	}
	delete(i.documents, id)
	return nil
}

func (i *Index) DeleteDocument(id core.DocumentID) error {
	i.docsMutex.Lock()
	defer i.docsMutex.Unlock()
	return i.deleteDocument(id)
}

func (i *Index) Length() int {
	i.docsMutex.RLock()
	defer i.docsMutex.RUnlock()
	return len(i.documents)
}

func (i *Index) GetDocument(id core.DocumentID) core.Document {
	i.docsMutex.RLock()
	defer i.docsMutex.RUnlock()
	return i.documents[id].Doc
}

func (i *Index) hasDocument(id core.DocumentID) bool {
	_, ok := i.documents[id]
	return ok
}

func (i *Index) GetAnalyzer() analysis.Analyzer {
	return i.analyzer
}

func (i *Index) GetTermFrequencies(id core.DocumentID) core.TermFrequency {
	return i.documents[id].Frequencies
}

func (i *Index) GetSpellSuggestions(terms []string, top int) (suggestions map[string][]string) {
	if i.fuzzyModel.Depth < 1 {
		return
	}
	suggestions = make(map[string][]string)
	for _, term := range terms {
		results := i.fuzzyModel.SpellCheckSuggestions(term, top)
		if results == nil {
			suggestions[term] = []string{}
		} else {
			suggestions[term] = results
		}
	}
	return
}

func (i *Index) filterDocuments(terms []string, filterFn search.Filter) map[core.DocumentID]core.DocumentData {
	i.docsMutex.RLock()
	defer i.docsMutex.RUnlock()
	docs := make(map[core.DocumentID]core.DocumentData)
	if filterFn == nil {
		for id, doc := range i.documents {
			docs[id] = doc
		}
	} else {
		docIds := filterFn(terms, i.terms)
		for _, id := range docIds {
			docs[id] = i.documents[id]
		}
	}
	return docs
}

func (i *Index) SearchDocuments(query string, filterFn search.Filter, rankFn search.Ranker) (documentScores []core.DocumentScore) {
	terms := i.analyzer.Analyze(query)
	docsData := i.filterDocuments(terms, filterFn)
	if rankFn == nil {
		for _, docData := range docsData {
			documentScores = append(documentScores, core.DocumentScore{Doc: docData.Doc, Score: 0})
		}
	} else {
		documentScores = rankFn(terms, docsData)
	}
	return
}
