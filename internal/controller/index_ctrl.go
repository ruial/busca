package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ruial/busca/internal/repository"
	"github.com/ruial/busca/pkg/analysis"
	"github.com/ruial/busca/pkg/index"
)

type IndexCtrl struct {
	IndexRepository repository.InMemoryIndexRepo
}

type IndexDTO struct {
	ID       string `json:"id" binding:"required"`
	Analyzer string `json:"analyzer" binding:"required"`
}

var analyzers = map[string]analysis.Analyzer{
	analysis.SimpleAnalyzer.String():     analysis.SimpleAnalyzer,
	analysis.WhiteSpaceAnalyzer.String(): analysis.WhiteSpaceAnalyzer,
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
	id := c.Param("id")
	idx, exists := ic.IndexRepository.GetIndex(id)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Index not found"})
		return
	}

	c.JSON(http.StatusOK, IndexDTO{ID: idx.ID, Analyzer: idx.Index.GetAnalyzer()})
}

func (ic IndexCtrl) DeleteIndex(c *gin.Context) {
	id := c.Param("id")
	if err := ic.IndexRepository.DeleteIndex(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
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
