# Busca

A minimal full text search engine developed in Go.

## Features

- RESTful API
- In-memory index with snapshot persistence
- Stop words and synonyms
- Snowball stemmer
- Multiple fast tokenizers
- tf-idf ranking with customizable weighting schemes
- Spell checking suggestions
- Cross platform and lightweight
- Observability with Prometheus metrics

## Usage

Check the [docs](docs) for API examples and my [blog](https://briefbytes.com) for more information about this project. To get started:

```sh
# API
docker-compose up
make bootstrap
curl -X POST http://localhost:8080/indexes/test-index/_search \
  -H 'content-type: application/json' \
  -d '{"query": "Alice dream adventures crime detective","filter": "or","tf_weight": "log","idf_weight": "smooth","include_text": false,"top": 3}'

# Library
idx := index.New(index.Opts{Analyzer: analysis.WhitespaceAnalyzer{}})
idx.AddDocument(core.NewBaseDocument("doc1", "some text"))
idx.AddDocument(core.NewBaseDocument("doc2", "even more text"))
ranker := search.TfIdfRanker(search.TfWeightLog, search.IdfWeightSmooth)
results := idx.SearchDocuments("more text", search.OrFilter, ranker)
fmt.Println(results)
```
