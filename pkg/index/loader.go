package index

import (
	"bytes"
	"encoding/gob"
	"io/ioutil"
	"log"
	"os"
	"path"
	"sync"

	"github.com/ruial/busca/pkg/analysis"
	"github.com/ruial/busca/pkg/core"
)

type gobber struct {
	index *Index
}

func (g *gobber) GobEncode() ([]byte, error) {
	g.index.docsMutex.Lock()
	defer g.index.docsMutex.Unlock()
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
	return w.Bytes(), nil
}

func (g *gobber) GobDecode(buf []byte) error {
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
	return decoder.Decode(&g.index.analyzer)
}

func loadFile(idx *Index, filePath string) error {
	bytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	document := core.NewBaseDocument(filePath, string(bytes))
	idx.AddDocument(&document)
	return nil
}

func LoadDocuments(dir string, analyzer analysis.Analyzer) (*Index, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	idx := New(analyzer)
	var wg sync.WaitGroup
	for _, file := range files {
		wg.Add(1)
		go func(file os.FileInfo) {
			defer wg.Done()
			filePath := path.Join(dir, file.Name())
			if err := loadFile(idx, filePath); err != nil {
				log.Printf("Unable to read file %s: %s", filePath, err.Error())
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
	g := &gobber{index: &Index{}}
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
	g := &gobber{index: index}
	if err = encoder.Encode(g); err != nil {
		return err
	}
	index = g.index
	return nil
}

func init() {
	// must register every type of interface implementations to export/import
	gob.Register(core.BaseDocument{})
	gob.Register(analysis.StandardAnalyzer)
	gob.Register(analysis.WhiteSpaceAnalyzer)
}
