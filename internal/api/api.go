package api

import (
	"github.com/gin-gonic/gin"
	"github.com/ruial/busca/internal/controller"
	"github.com/ruial/busca/internal/repository"
)

func Server(addr string) error {
	router := gin.Default()

	indexController := controller.IndexCtrl{IndexRepository: repository.NewInMemoryIndexRepo()}

	router.GET("/index", indexController.GetIndexes)
	router.POST("/index", indexController.CreateIndex)
	router.GET("/index/:id", indexController.GetIndex)
	router.DELETE("/index/:id", indexController.DeleteIndex)

	return router.Run(addr)
}
