package controller

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

type IndexDTO struct {
	ID       string `json:"id" binding:"required"`
	Analyzer string `json:"analyzer" binding:"required"`
}

type DocumentDTO struct {
	ID   string `json:"id" binding:"required"`
	Text string `json:"text" binding:"required"`
}

type DocumentUpdateDTO struct {
	Text string `json:"text" binding:"required"`
}

type SearchInputDTO struct {
	Query     string
	Filter    string
	MinMatch  string `json:"min_match"`
	TfWeight  string `json:"tf_weight"`
	IdfWeight string `json:"idf_weight"`
}

type DocumentScoreDTO struct {
	ID    string  `json:"id"`
	Score float64 `json:"score"`
}
type SearchOutputDTO struct {
	Docs []DocumentScoreDTO `json:"docs"`
	Size int                `json:"size"`
}

var analyzers = map[string]analysis.Analyzer{
	analysis.StandardAnalyzer.String():   analysis.StandardAnalyzer,
	analysis.WhiteSpaceAnalyzer.String(): analysis.WhiteSpaceAnalyzer,
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

func (ic IndexCtrl) index(c *gin.Context) *index.Index {
	idx, _ := c.Get("index")
	return idx.(*index.Index)
}

func (ic IndexCtrl) CreateIndex(c *gin.Context) {
	var json IndexDTO
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	analyzer, ok := analyzers[json.Analyzer]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Analyzer not found"})
		return
	}

	idx := repository.IdentifiableIndex{ID: json.ID, Index: index.New(analyzer)}
	if err := ic.IndexRepository.CreateIndex(idx); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, json)
}

func (ic IndexCtrl) GetIndex(c *gin.Context) {
	idx := ic.index(c)
	c.JSON(http.StatusOK, IndexDTO{ID: c.Param("id"), Analyzer: idx.GetAnalyzer()})
}

func (ic IndexCtrl) DeleteIndex(c *gin.Context) {
	if err := ic.IndexRepository.DeleteIndex(c.Param("id")); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	}
	c.Status(http.StatusNoContent)
}

func (ic IndexCtrl) GetIndexes(c *gin.Context) {
	idxs := ic.IndexRepository.GetIndexes()
	list := make([]IndexDTO, 0, len(idxs))

	for _, idx := range idxs {
		list = append(list, IndexDTO{ID: idx.ID, Analyzer: idx.Index.GetAnalyzer()})
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
	if err := idx.AddDocument(core.NewBaseDocument(json.ID, json.Text)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, json)
}

func (ic IndexCtrl) GetDocument(c *gin.Context) {
	idx := ic.index(c)
	doc := idx.GetDocument(c.Param("docId"))
	if doc == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": index.ErrNonExistentDocument.Error()})
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
	updateFunction := idx.UpdateDocument
	upsert := strings.ToLower(c.Query("upsert"))
	if upsert == "true" {
		updateFunction = idx.UpsertDocument
	}

	doc := core.NewBaseDocument(c.Param("docId"), json.Text)
	if err := updateFunction(doc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, DocumentDTO{ID: doc.ID(), Text: doc.Text()})
}

func (ic IndexCtrl) DeleteDocument(c *gin.Context) {
	idx := ic.index(c)
	if err := idx.DeleteDocument(c.Param("docId")); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (ic IndexCtrl) SearchDocuments(c *gin.Context) {
	dto := SearchInputDTO{
		Filter:    "or",
		TfWeight:  "default",
		IdfWeight: "default",
	}
	if c.Request.Method == http.MethodGet {
		dto.Query = c.Query("query")
		dto.Filter = c.DefaultQuery("filter", dto.Filter)
		dto.MinMatch = c.Query("min_match")
		dto.TfWeight = c.DefaultQuery("tf_weight", dto.TfWeight)
		dto.IdfWeight = c.DefaultQuery("idf_weight", dto.IdfWeight)
	} else {
		if err := c.ShouldBindJSON(&dto); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	filter, ok := filters[dto.Filter]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid filter"})
		return
	}

	if dto.MinMatch != "" {
		minMatch, err := strconv.Atoi(dto.MinMatch)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid min_match"})
			return
		}
		filter = search.MinMatchFilter(minMatch)
	}

	tfWeight, ok := tfWeights[dto.TfWeight]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tf weight"})
		return
	}
	idfWeight, ok := idfWeights[dto.IdfWeight]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid idf weight"})
		return
	}

	ranker := search.TfIdfRanker(tfWeight, idfWeight)
	idx := ic.index(c)
	docs := idx.SearchDocuments(dto.Query, filter, ranker)

	docsDto := make([]DocumentScoreDTO, 0, len(docs))
	for _, res := range docs {
		docsDto = append(docsDto, DocumentScoreDTO{ID: res.Doc.ID(), Score: res.Score})
	}

	c.JSON(http.StatusOK, SearchOutputDTO{Docs: docsDto, Size: len(docs)})
}
