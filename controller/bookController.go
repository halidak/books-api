package controller

import (
	"context"
	"example/books-api/model"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var bookCollection *mongo.Collection
var readingListCollection *mongo.Collection

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	connectionString := os.Getenv("CONNECTION_STRING")
	clientOptions := options.Client().ApplyURI(connectionString)

	client, error := mongo.Connect(context.TODO(), clientOptions)

	if error != nil {
		log.Fatal(error)
	}

	fmt.Println("Mongodb connection success")

	dbName := os.Getenv("DBNAME")
    colName2 := os.Getenv("COLNAME2")
	colName := os.Getenv("COLNAME")

	bookCollection = client.Database(dbName).Collection(colName2)
	readingListCollection = client.Database(dbName).Collection(colName)

	fmt.Println("Collection istance is ready")
}

//insert book with author
func insertBook(book model.Book) {
	inserted, err := bookCollection.InsertOne(context.Background(), book)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Inserted a single document: ", inserted.InsertedID)

	authorID := book.Author
    update := bson.M{"$addToSet": bson.M{"books": inserted.InsertedID}}
    filter := bson.M{"_id": authorID}

    _, err = readingListCollection.UpdateOne(context.Background(), filter, update)
    if err != nil {
        log.Fatal(err)
    }
}

//get book with author name
func getBookWithAuthor(bookId string) (model.BookWithAuthor, error) {
    var bookWithAuthor model.BookWithAuthor

    id, err := primitive.ObjectIDFromHex(bookId)
    if err != nil {
        return bookWithAuthor, err
    }

    pipeline := []bson.M{
        {"$match": bson.M{"_id": id}},
        {"$lookup": bson.M{
            "from":         "readList",
            "localField":   "author",
            "foreignField": "_id",
            "as":           "authorInfo",
        }},
        {"$unwind": "$authorInfo"},
        {"$project": bson.M{
            "_id":    1,
            "title":  1,
            "genre":  1,
            "author": bson.M{"_id": "$authorInfo._id", "name": "$authorInfo.name"},
            "read":   1,
        }},
    }

    cursor, err := bookCollection.Aggregate(context.Background(), pipeline)
    if err != nil {
        log.Fatal(err)
    }

    if cursor.Next(context.Background()) {
        cursor.Decode(&bookWithAuthor)
    }

    return bookWithAuthor, nil
}


//get all book with author name
func getAllBooksWithAuthors() []model.BookWithAuthor {
    var booksWithAuthors []model.BookWithAuthor

    pipeline := []bson.M{
        {"$lookup": bson.M{
            "from":         "readList",
            "localField":   "author",
            "foreignField": "_id",
            "as":           "authorInfo",
        }},
        {"$unwind": "$authorInfo"},
        {"$project": bson.M{
            "_id":    1,
            "title":  1,
            "genre":  1,
            "author": bson.M{"_id": "$authorInfo._id", "name": "$authorInfo.name"},
            "read":   1,
        }},
    }

    cursor, err := bookCollection.Aggregate(context.Background(), pipeline)

    if err != nil {
        log.Fatal(err)
    }

    for cursor.Next(context.Background()) {
        var bookWithAuthor model.BookWithAuthor
        cursor.Decode(&bookWithAuthor)
        booksWithAuthors = append(booksWithAuthors, bookWithAuthor)
    }

    return booksWithAuthors
}

//update book
func updateBook(bookId string, book model.Book) {
	id, _ := primitive.ObjectIDFromHex(bookId)
	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{"title": book.Title, "genre": book.Genre, "author": book.Author, "read": book.Read}}

	result, err := bookCollection.UpdateOne(context.Background(), filter, update)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Updated a single document: ", result.UpsertedID)
}

//delete book
func deleteBook(bookId string) {
	id, _ := primitive.ObjectIDFromHex(bookId)
	filter := bson.M{"_id": id}

	result, err := bookCollection.DeleteOne(context.Background(), filter)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Deleted a single document: ", result.DeletedCount)
}

func GetAllBooksWithAuthors(c *gin.Context) {
	allBooksWithAuthors := getAllBooksWithAuthors()
	c.JSON(http.StatusOK, allBooksWithAuthors)
}

func GetBookWithAuthor(c *gin.Context) {
	bookId := c.Param("bookId")
	bookWithAuthor, err := getBookWithAuthor(bookId)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Book not found"})
		return
	}
	c.JSON(http.StatusOK, bookWithAuthor)
}

func CreateBook(c *gin.Context) {
	var book model.Book
	if err := c.ShouldBindJSON(&book); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	insertBook(book)
	c.JSON(http.StatusOK, book)
}

func UpdateBook(c *gin.Context) {
	bookId := c.Param("bookId")
	var book model.Book
	if err := c.ShouldBindJSON(&book); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updateBook(bookId, book)
	c.JSON(http.StatusOK, book)
}

func DeleteBook(c *gin.Context) {
	bookId := c.Param("bookId")
	deleteBook(bookId)
	c.JSON(http.StatusOK, gin.H{"message": "Book deleted"})
}

//get all books from author
func getAllBooksForAuthor(authorId primitive.ObjectID) []model.Book {
    var books []model.Book

    pipeline := []bson.M{
        {"$match": bson.M{"author": authorId}},
        {"$lookup": bson.M{
            "from":         "readList",
            "localField":   "author",
            "foreignField": "_id",
            "as":           "authorInfo",
        }},
        {"$unwind": "$authorInfo"},
        {"$project": bson.M{
            "_id":    1,
            "title":  1,
            "genre":  1,
            "author": "$authorInfo.name",
            "read":   1,
        }},
    }

    cursor, err := bookCollection.Aggregate(context.Background(), pipeline)
    if err != nil {
        log.Fatal(err)
    }

    for cursor.Next(context.Background()) {
        var book model.Book
        cursor.Decode(&book)
        books = append(books, book)
    }

    return books
}

//get all books from author
func GetBooksForAuthor(c *gin.Context) {
    authorId := c.Param("authorId")

    // Convert authorId to ObjectID
    objAuthorId, err := primitive.ObjectIDFromHex(authorId)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid author ID"})
        return
    }

    // Call the function to get all books for the author
    booksForAuthor := getAllBooksForAuthor(objAuthorId)

    // Return the result as JSON
    c.JSON(http.StatusOK, booksForAuthor)
}

//read book
func readBook(bookId string) {
	id, _ := primitive.ObjectIDFromHex(bookId)
	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{"read": true}}

	result, err := bookCollection.UpdateOne(context.Background(), filter, update)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Updated a single document: ", result.UpsertedID)
}

func ReadBook(c *gin.Context) {
	bookId := c.Param("bookId")
	readBook(bookId)
	c.JSON(http.StatusOK, gin.H{"message": "Book read"})
}