package index

import (
	"bytes"
	"encoding/gob"
	"log"
	"os"
	"path"
	"sync"

	"github.com/ruial/busca/pkg/analysis"
	"github.com/ruial/busca/pkg/core"
	"github.com/sajari/fuzzy"
)

type Gobber struct {
	index *Index
}

// fuzzy.Model has an embedded mutex, which causes issues with GobEncoder
type fuzzyEncodable struct {
	Data      map[string]*fuzzy.Counts
	Maxcount  int
	Suggest   map[string][]string
	Depth     int
	Threshold int
}

func (g *Gobber) GobEncode() ([]byte, error) {
	g.index.docsMutex.RLock()
	defer g.index.docsMutex.RUnlock()
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	if err := encoder.Encode(g.index.terms); err != nil {
		return nil, err
	}
	if err := encoder.Encode(g.index.documents); err != nil {
		return nil, err
	}
	if err := encoder.Encode(&g.index.analyzer); err != nil {
		return nil, err
	}
	// assume that Index was initialized with New, so fuzzyModel is never nil
	fuzzyModel := fuzzyEncodable{
		Data:      g.index.fuzzyModel.Data,
		Maxcount:  g.index.fuzzyModel.Maxcount,
		Suggest:   g.index.fuzzyModel.Suggest,
		Depth:     g.index.fuzzyModel.Depth,
		Threshold: g.index.fuzzyModel.Threshold,
	}
	if err := encoder.Encode(&fuzzyModel); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func (g *Gobber) GobDecode(buf []byte) error {
	g.index.docsMutex.Lock()
	defer g.index.docsMutex.Unlock()
	r := bytes.NewBuffer(buf)
	decoder := gob.NewDecoder(r)
	if err := decoder.Decode(&g.index.terms); err != nil {
		return err
	}
	if err := decoder.Decode(&g.index.documents); err != nil {
		return err
	}
	if err := decoder.Decode(&g.index.analyzer); err != nil {
		return err
	}
	fuzzyModel := &fuzzyEncodable{}
	if err := decoder.Decode(fuzzyModel); err != nil {
		return err
	}
	g.index.fuzzyModel = &fuzzy.Model{
		Data:      fuzzyModel.Data,
		Maxcount:  fuzzyModel.Maxcount,
		Suggest:   fuzzyModel.Suggest,
		Depth:     fuzzyModel.Depth,
		Threshold: fuzzyModel.Threshold,
	}
	return nil
}

func loadFile(idx *Index, filePath string) error {
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	document := core.NewBaseDocument(filePath, string(bytes))
	idx.AddDocument(&document)
	return nil
}

func LoadDocuments(dir string, indexOpts Opts) (*Index, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	idx := New(indexOpts)
	var wg sync.WaitGroup
	for _, file := range files {
		wg.Add(1)
		go func(file os.DirEntry) {
			defer wg.Done()
			if !file.IsDir() {
				filePath := path.Join(dir, file.Name())
				if err := loadFile(idx, filePath); err != nil {
					log.Printf("Unable to read file %s: %s", filePath, err.Error())
				}
			}
		}(file)
	}
	wg.Wait()
	return idx, nil
}

func Import(path string) (*Index, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	decoder := gob.NewDecoder(file)
	g := &Gobber{index: &Index{}}
	err = decoder.Decode(g)
	return g.index, err
}

func Export(index *Index, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := gob.NewEncoder(file)
	g := &Gobber{index: index}
	if err = encoder.Encode(g); err != nil {
		return err
	}
	index = g.index
	return nil
}

func init() {
	// must register every type of interface implementations to export/import
	gob.Register(core.BaseDocument{})
	gob.Register(analysis.StandardAnalyzer{})
	gob.Register(analysis.SimpleAnalyzer{})
	gob.Register(analysis.WhitespaceAnalyzer{})
}
