package controller

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ruial/busca/internal/repository"
	"github.com/ruial/busca/pkg/analysis"
	"github.com/ruial/busca/pkg/core"
	"github.com/ruial/busca/pkg/index"
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

var analyzers = map[string]analysis.Analyzer{
	analysis.SimpleAnalyzer.String():     analysis.SimpleAnalyzer,
	analysis.WhiteSpaceAnalyzer.String(): analysis.WhiteSpaceAnalyzer,
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
