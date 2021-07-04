#!/bin/sh
# Script to compare busca and elasticsearch index/search speeds
# Search results are a slightly different because of the ranking algorithms: tf-idf and bm25
# As busca is much simpler and has less features, it is faster on a single node

curl -X POST http://localhost:8080/indexes \
  -H 'Content-Type: application/json; charset=utf-8' \
  -d'{"id": "test-index","analyzer": "WhitespaceAnalyzer"}'

curl -X PUT http://localhost:9200/test-index \
  -H 'Content-Type: application/json; charset=utf-8' \
  -d'{"settings": {"number_of_replicas": 0},"mappings": {"properties": {"text": { "type": "text", "analyzer": "whitespace" }}}}'

go run examples/loader/main.go

# do 100 searches with a max of 20 in parallel
time seq 1 100 | xargs -n1 -P20 curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:8080/indexes/test-index/_search \
  -H 'content-type: application/json' \
  -d'{"query": "adventures of a smart crime detective","filter": "or","tf_weight": "log","idf_weight": "smooth","include_text": false}'

time seq 1 100 | xargs -n1 -P20 curl -s -o /dev/null -w "%{http_code}" -X POST 'http://localhost:9200/test-index/_search?size=100&filter_path=hits.total.value%2Chits.hits._id%2Chits.hits._score' \
  -H 'content-type: application/json' \
  -d'{"query": {"match": {"text": {"query": "adventures of a smart crime detective","operator": "or"}}}}'

#curl -X DELETE http://localhost:8080/indexes/test-index

#curl -X DELETE http://localhost:9200/test-index
