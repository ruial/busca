package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ruial/busca/internal/repository"
	"github.com/ruial/busca/pkg/analysis"
	"github.com/ruial/busca/pkg/core"
	"github.com/ruial/busca/pkg/index"
)

var router *gin.Engine

func init() {
	indexRepo := &repository.LocalIndexRepo{}
	index := index.New(index.Opts{Analyzer: analysis.SimpleAnalyzer{}})
	index.AddDocument(core.NewBaseDocument("doc1", "sample document"))
	indexRepo.CreateIndex(repository.IdentifiableIndex{ID: "test", Index: index})
	router = SetupRouter(indexRepo)
}

func TestCreateIndex(t *testing.T) {
	indexDTO := IndexInputDTO{ID: "test2", Analyzer: "WhitespaceAnalyzer"}
	indexJSON, _ := json.Marshal(indexDTO)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/indexes", bytes.NewBuffer(indexJSON))
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Error("Index create should have status 201, got:", w.Code)
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/indexes", bytes.NewBuffer(indexJSON))
	router.ServeHTTP(w, req)
	if w.Code != http.StatusConflict {
		t.Error("Repeated index should get status 409, got:", w.Code)
	}
}

func TestGetDocumentDoesNotExist(t *testing.T) {
	// document exists
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/indexes/test/docs/doc1", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Error("Status code should have been ok, got:", w.Code)
	}

	// index does not exist
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/indexes/test0/docs/doc1", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Error("Status code should have been not found, got:", w.Code)
	}

	if !strings.Contains(w.Body.String(), repository.ErrIndexDoesNotExist.Error()) {
		t.Error("Should get message of index does not exist")
	}

	// document does not exist
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/indexes/test/docs/doc2", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Error("Status code should have been not found, got:", w.Code)
	}

	if !strings.Contains(w.Body.String(), index.ErrNonExistentDocument.Error()) {
		t.Error("Should get message of document does not exist")
	}
}
