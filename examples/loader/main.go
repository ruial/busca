package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/ruial/busca/pkg/analysis"
	"github.com/ruial/busca/pkg/core"
	"github.com/ruial/busca/pkg/index"
)

func addDocument(endpoint string, doc core.Document) error {
	client := &http.Client{Timeout: 5 * time.Second}
	docJSON, err := json.Marshal(doc)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPut, endpoint, bytes.NewBuffer(docJSON))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if !(resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated) {
		return errors.New(resp.Status)
	}
	_, err = io.ReadAll(resp.Body)
	return err
}

func addDocEndpoint(engine, idxId, id string) string {
	if engine == "busca" {
		return fmt.Sprintf("http://localhost:8080/indexes/%s/docs/%s?upsert=true", idxId, id)
	} else if engine == "elastic" {
		return fmt.Sprintf("http://localhost:9200/%s/_doc/%s", idxId, id)
	}
	panic("unsupported engine")
}

func addDocuments(engine string, idxId string, docs []core.DocumentScore) {
	start := time.Now()
	var wg sync.WaitGroup
	for _, res := range docs {
		wg.Add(1)
		parts := strings.Split(res.Doc.ID(), "/")
		id := parts[len(parts)-1]
		// fmt.Println("adding doc", i, id)

		go func(res core.DocumentScore) {
			defer wg.Done()
			endpoint := addDocEndpoint(engine, idxId, id)
			if err := addDocument(endpoint, core.NewBaseDocument(id, res.Doc.Text())); err != nil {
				panic(err)
			}
		}(res)
	}
	wg.Wait()
	fmt.Println("added:", len(docs), "elapsed:", time.Now().Sub(start))
}

func main() {
	// Measure index speed on busca and elasticsearch on a sample dataset
	fmt.Println("loading documents to in memory index")
	idx, err := index.LoadDocuments("testdata/books", analysis.WhitespaceAnalyzer{})
	if err != nil {
		panic(err)
	}

	indexId := "test-index"
	docs := idx.SearchDocuments("", nil, nil)
	fmt.Println("adding docs on busca")
	addDocuments("busca", indexId, docs)
	// fmt.Println("adding docs on elastic")
	// addDocuments("elastic", indexId, docs)
}
