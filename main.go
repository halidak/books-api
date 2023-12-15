package main

import (
	"example/books-api/router"
	"fmt"
	"log"
	"net/http"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }
	fmt.Println("Server is getting started...")
	r := router.AuthorRoutes()
	log.Fatal(http.ListenAndServe(":8000", r))
	fmt.Println("Listening on port 8000")
}