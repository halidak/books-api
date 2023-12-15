package router

import (
	"example/books-api/controller"

	"github.com/gin-gonic/gin"
)

func AuthorRoutes() *gin.Engine {
	router := gin.Default()

	authorGroup := router.Group("/author")
	{
		authorGroup.POST("/add", controller.CreateAuthor)
		authorGroup.GET("/all", controller.GetAllAuthors)
		authorGroup.GET("/:authorId", controller.GetAuthor)
		authorGroup.PUT("/:authorId", controller.UpdateAuthor)
		authorGroup.DELETE("/:authorId", controller.DeleteAuthor)
	}

	return router
}
