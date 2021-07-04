package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ruial/busca/internal/repository"
)

func indexExists(ic IndexCtrl) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		idx, exists := ic.IndexRepository.GetIndex(id)
		if !exists {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": repository.ErrIndexDoesNotExist.Error()})
			return
		}
		c.Set("index", idx)
	}
}

func SetupRouter(indexRepository repository.IndexRepo) *gin.Engine {
	router := gin.Default()

	indexController := IndexCtrl{IndexRepository: indexRepository}
	indexExistsMiddleware := indexExists(indexController)

	router.GET("/indexes", indexController.GetIndexes)
	router.POST("/indexes", indexController.CreateIndex)
	router.GET("/indexes/:id", indexExistsMiddleware, indexController.GetIndex)
	router.DELETE("/indexes/:id", indexExistsMiddleware, indexController.DeleteIndex)

	router.GET("/indexes/:id/_search", indexExistsMiddleware, indexController.SearchDocuments)
	router.POST("/indexes/:id/_search", indexExistsMiddleware, indexController.SearchDocuments)

	router.POST("/indexes/:id/docs", indexExistsMiddleware, indexController.CreateDocument)
	router.GET("/indexes/:id/docs/:docId", indexExistsMiddleware, indexController.GetDocument)
	router.PUT("/indexes/:id/docs/:docId", indexExistsMiddleware, indexController.UpdateDocument)
	router.DELETE("/indexes/:id/docs/:docId", indexExistsMiddleware, indexController.DeleteDocument)

	return router
}
