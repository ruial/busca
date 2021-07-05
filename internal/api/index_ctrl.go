package api

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ruial/busca/internal/repository"
	"github.com/ruial/busca/pkg/analysis"
	"github.com/ruial/busca/pkg/core"
	"github.com/ruial/busca/pkg/index"
	"github.com/ruial/busca/pkg/search"
)

type IndexCtrl struct {
	IndexRepository repository.IndexRepo
}

type IndexInputDTO struct {
	ID        string            `json:"id" binding:"required"`
	Analyzer  string            `json:"analyzer" binding:"required"`
	Stopwords []string          `json:"stopwords,omitempty"`
	Synonyms  map[string]string `json:"synonyms,omitempty"`
}

type IndexOutputDTO struct {
	ID        string            `json:"id"`
	Analyzer  string            `json:"analyzer"`
	Docs      int               `json:"docs"`
	Stopwords []string          `json:"stopwords,omitempty"`
	Synonyms  map[string]string `json:"synonyms,omitempty"`
}

type DocumentDTO struct {
	ID   string `json:"id" binding:"required"`
	Text string `json:"text" binding:"required"`
}

type DocumentUpdateDTO struct {
	Text string `json:"text" binding:"required"`
}

type SearchInputDTO struct {
	Query       string
	Filter      string
	MinMatch    *int   `json:"min_match"`
	TfWeight    string `json:"tf_weight"`
	IdfWeight   string `json:"idf_weight"`
	IncludeText bool   `json:"include_text"`
}

type DocumentScoreDTO struct {
	ID    string  `json:"id"`
	Text  string  `json:"text,omitempty"`
	Score float64 `json:"score"`
}

type SearchOutputDTO struct {
	Docs []DocumentScoreDTO `json:"docs"`
	Size int                `json:"size"`
}

func newAnalyzer(dto IndexInputDTO) analysis.Analyzer {
	analyzer := strings.ToLower(dto.Analyzer)
	stopwords := make(map[string]struct{})
	for _, stopword := range dto.Stopwords {
		stopwords[stopword] = struct{}{}
	}
	if analyzer == "standardanalyzer" {
		return analysis.StandardAnalyzer{Settings: analysis.Settings{Stopwords: stopwords, Synonyms: dto.Synonyms}}
	}
	if analyzer == "whitespaceanalyzer" {
		return analysis.WhitespaceAnalyzer{Settings: analysis.Settings{Stopwords: stopwords, Synonyms: dto.Synonyms}}
	}
	return nil
}

var filters = map[string]search.Filter{
	"and": search.AndFilter,
	"or":  search.OrFilter,
}

var tfWeights = map[string]search.TfWeightScheme{
	"default": search.TfWeightDefault,
	"log":     search.TfWeightLog,
}

var idfWeights = map[string]search.IdfWeightScheme{
	"default": search.IdfWeightDefault,
	"smooth":  search.IdfWeightSmooth,
}

func (ic IndexCtrl) index(c *gin.Context) repository.IdentifiableIndex {
	idx, _ := c.Get("index")
	return idx.(repository.IdentifiableIndex)
}

func (ic IndexCtrl) CreateIndex(c *gin.Context) {
	var json IndexInputDTO
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	analyzer := newAnalyzer(json)
	if analyzer == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Analyzer not found"})
		return
	}

	idx, err := repository.NewIdentifiableIndex(json.ID, index.New(analyzer))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := ic.IndexRepository.CreateIndex(idx); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, json)
}

func (ic IndexCtrl) GetIndex(c *gin.Context) {
	idx := ic.index(c).Index
	analyzer := idx.GetAnalyzer()
	c.JSON(http.StatusOK,
		IndexOutputDTO{ID: c.Param("id"), Analyzer: analyzer.String(),
			Docs: idx.Length(), Stopwords: analyzer.GetStopwords(), Synonyms: analyzer.GetSynonyms()})
}

func (ic IndexCtrl) DeleteIndex(c *gin.Context) {
	if err := ic.IndexRepository.DeleteIndex(c.Param("id")); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (ic IndexCtrl) GetIndexes(c *gin.Context) {
	indexes := ic.IndexRepository.GetIndexes()
	list := make([]IndexOutputDTO, 0, len(indexes))

	for _, idx := range indexes {
		outputDto := IndexOutputDTO{ID: idx.ID, Analyzer: idx.Index.GetAnalyzer().String(), Docs: idx.Index.Length()}
		list = append(list, outputDto)
	}

	c.JSON(http.StatusOK, gin.H{"indexes": list})
}

func (ic IndexCtrl) CreateDocument(c *gin.Context) {
	var json DocumentDTO
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	idx := ic.index(c)
	if err := idx.Index.AddDocument(core.NewBaseDocument(json.ID, json.Text)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, json)
}

func (ic IndexCtrl) GetDocument(c *gin.Context) {
	idx := ic.index(c)
	doc := idx.Index.GetDocument(c.Param("docId"))
	if doc == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": index.ErrNonExistentDocument.Error()})
		return
	}
	c.JSON(http.StatusOK, DocumentDTO{ID: doc.ID(), Text: doc.Text()})
}

func (ic IndexCtrl) UpdateDocument(c *gin.Context) {
	var json DocumentUpdateDTO
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	idx := ic.index(c)
	updateFunction := idx.Index.UpdateDocument
	upsert := strings.ToLower(c.Query("upsert"))
	if upsert == "true" {
		updateFunction = idx.Index.UpsertDocument
	}

	doc := core.NewBaseDocument(c.Param("docId"), json.Text)
	if err := updateFunction(doc); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, DocumentDTO{ID: doc.ID(), Text: doc.Text()})
}

func (ic IndexCtrl) DeleteDocument(c *gin.Context) {
	idx := ic.index(c)
	if err := idx.Index.DeleteDocument(c.Param("docId")); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (ic IndexCtrl) SearchDocuments(c *gin.Context) {
	inputDto := SearchInputDTO{
		Filter:      "or",
		TfWeight:    "default",
		IdfWeight:   "default",
		IncludeText: true,
	}
	if c.Request.Method == http.MethodGet {
		inputDto.Query = c.Query("query")
		inputDto.Filter = c.DefaultQuery("filter", inputDto.Filter)
		minMatchStr := c.Query("min_match")
		if minMatchStr != "" {
			minMatch, err := strconv.Atoi(minMatchStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid min_match"})
				return
			}
			inputDto.MinMatch = &minMatch
		}
		inputDto.TfWeight = c.DefaultQuery("tf_weight", inputDto.TfWeight)
		inputDto.IdfWeight = c.DefaultQuery("idf_weight", inputDto.IdfWeight)
		if strings.ToLower(c.Query("include_text")) == "false" {
			inputDto.IncludeText = false
		}
	} else {
		if err := c.ShouldBindJSON(&inputDto); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	filter, ok := filters[inputDto.Filter]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid filter"})
		return
	}
	// if minMatch is defined and it is <= 0, filter will get all documents
	if inputDto.MinMatch != nil {
		filter = search.MinMatchFilter(*inputDto.MinMatch)
	}

	tfWeight, ok := tfWeights[inputDto.TfWeight]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tf weight"})
		return
	}
	idfWeight, ok := idfWeights[inputDto.IdfWeight]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid idf weight"})
		return
	}

	ranker := search.TfIdfRanker(tfWeight, idfWeight)
	idx := ic.index(c).Index
	docs := idx.SearchDocuments(inputDto.Query, filter, ranker)

	docsDto := make([]DocumentScoreDTO, 0, len(docs))
	for _, res := range docs {
		docDto := DocumentScoreDTO{ID: res.Doc.ID(), Score: res.Score}
		if inputDto.IncludeText {
			docDto.Text = res.Doc.Text()
		}
		docsDto = append(docsDto, docDto)
	}

	c.JSON(http.StatusOK, SearchOutputDTO{Docs: docsDto, Size: len(docs)})
}

func (ic IndexCtrl) Analyze(c *gin.Context) {
	idx := ic.index(c)
	analyzer := idx.Index.GetAnalyzer()
	tokens := analyzer.Analyze(c.Query("text"))
	c.JSON(http.StatusOK, tokens)
}
