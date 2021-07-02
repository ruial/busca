package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ruial/busca/internal/controller"
	"github.com/ruial/busca/internal/repository"
)

func indexExists(ic controller.IndexCtrl) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		idx, exists := ic.IndexRepository.GetIndex(id)
		if !exists {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": repository.ErrIndexDoesNotExist.Error()})
			return
		}
		c.Set("index", idx.Index)
	}
}

func Server(addr string, indexRepository repository.IndexRepo) error {
	router := gin.Default()

	indexController := controller.IndexCtrl{IndexRepository: indexRepository}
	indexExistsMiddleware := indexExists(indexController)

	router.GET("/index", indexController.GetIndexes)
	router.POST("/index", indexController.CreateIndex)
	router.GET("/index/:id", indexExistsMiddleware, indexController.GetIndex)
	router.DELETE("/index/:id", indexExistsMiddleware, indexController.DeleteIndex)

	router.POST("/index/:id/", indexExistsMiddleware, indexController.CreateDocument)
	router.GET("/index/:id/:docId", indexExistsMiddleware, indexController.GetDocument)
	router.PUT("/index/:id/:docId", indexExistsMiddleware, indexController.UpdateDocument)
	router.DELETE("/index/:id/:docId", indexExistsMiddleware, indexController.DeleteDocument)

	return router.Run(addr)
}
