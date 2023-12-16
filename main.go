package main

import (
	"example/books-api/router"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	fmt.Println("Server is getting started...")

	r := gin.Default()

	// Register both author and book routes
	authorRoutes := router.AuthorRoutes()
	bookRoutes := router.BookRoutes()

	// Combine the routes using groups
	authorGroup := r.Group("/author")
	authorGroup.Any("/*path", gin.WrapH(authorRoutes))

	bookGroup := r.Group("/book")
	bookGroup.Any("/*path", gin.WrapH(bookRoutes))

	// Start the server
	log.Fatal(http.ListenAndServe(":8000", r))
	fmt.Println("Listening on port 8000")
}
