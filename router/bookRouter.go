package router

import (
	"example/books-api/controller"
	"github.com/gin-gonic/gin"
)

func BookRoutes() *gin.Engine {
	router := gin.Default()

	bookGroup := router.Group("/book")
	{
		bookGroup.POST("/add", controller.CreateBook)
		bookGroup.GET("/all", controller.GetAllBooksWithAuthors)
		bookGroup.GET("/:bookId", controller.GetBookWithAuthor)
		bookGroup.GET("/author-books/:authorId", controller.GetBooksForAuthor)
		bookGroup.PUT("/read-book/:bookId", controller.ReadBook)
		bookGroup.PUT("/:bookId", controller.UpdateBook)
		bookGroup.DELETE("/:bookId", controller.DeleteBook)
	}

	return router
}
