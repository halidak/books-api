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

var collection *mongo.Collection
var bookAuthorCollection *mongo.Collection
var bookListCollection *mongo.Collection

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
    colName := os.Getenv("COLNAME")
    colName2 := os.Getenv("COLNAME2")
    colName3 := os.Getenv("COLNAME3")

	collection = client.Database(dbName).Collection(colName)
    bookListCollection = client.Database(dbName).Collection(colName2)
    bookAuthorCollection = client.Database(dbName).Collection(colName3)

	fmt.Println("Collection istance is ready")
}

// insert author
func insertAuthor(author model.Author) {
	inserted, err := collection.InsertOne(context.Background(), author)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Inserted a single document: ", inserted.InsertedID)
}

// update author
func updateAuthor(authorId string, author model.Author) {
    id, _ := primitive.ObjectIDFromHex(authorId)
    filter := bson.M{"_id": id}
    update := bson.M{"$set": bson.M{"name": author.Name}} 

    result, err := collection.UpdateOne(context.Background(), filter, update)

    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Updated a single document: ", result.UpsertedID)
}

// delete author
func deleteAuthor(authorId string) {
    id, _ := primitive.ObjectIDFromHex(authorId)
    filter := bson.M{"_id": id}

    // Get the author's books from the bookAuthor table
    bookAuthorFilter := bson.M{"author": id}
    cursor, err := bookAuthorCollection.Find(context.Background(), bookAuthorFilter)
    if err != nil {
        log.Fatal(err)
    }

    var bookIds []primitive.ObjectID
    for cursor.Next(context.Background()) {
        var bookAuthor model.BookAuthor
        cursor.Decode(&bookAuthor)
        bookIds = append(bookIds, bookAuthor.Book)
    }

    result, err := collection.DeleteOne(context.Background(), filter)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Deleted a single document from authors: ", result.DeletedCount)

    bookAuthorResult, err := bookAuthorCollection.DeleteMany(context.Background(), bookAuthorFilter)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Deleted documents from bookAuthor: ", bookAuthorResult.DeletedCount)

    readListFilter := bson.M{"book": bson.M{"$in": bookIds}}
    readListResult, err := collection.DeleteMany(context.Background(), readListFilter)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Deleted documents from readList: ", readListResult.DeletedCount)

    bookFilter := bson.M{"_id": bson.M{"$in": bookIds}}
    bookResult, err := bookListCollection.DeleteMany(context.Background(), bookFilter)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Deleted documents from bookList: ", bookResult.DeletedCount)
}

// get author and return
func getAuthor(authorID string) (model.AuthorWithBooks, error) {
    var authorWithBooks model.AuthorWithBooks

    id, err := primitive.ObjectIDFromHex(authorID)
    if err != nil {
        return authorWithBooks, err
    }

    pipeline := []bson.M{
        {"$match": bson.M{"_id": id}},
        {"$lookup": bson.M{
            "from":         "bookAuthor",
            "localField":   "_id",
            "foreignField": "author",
            "as":           "authorBookRelations",
        }},
        {"$lookup": bson.M{
            "from":         "bookList",
            "localField":   "authorBookRelations.book",
            "foreignField": "_id",
            "as":           "books",
        }},
        {"$project": bson.M{
            "_id":   1,
            "name":  1,
            "books": bson.M{"$ifNull": []interface{}{
                bson.M{"$map": bson.M{
                    "input": "$books",
                    "as":    "book",
                    "in": bson.M{"title": "$$book.title"},
                }},
                []model.BookInfo{},
            }},
        }},
    }

    cursor, err := collection.Aggregate(context.Background(), pipeline)
    if err != nil {
        log.Fatal(err)
    }

    if cursor.Next(context.Background()) {
        cursor.Decode(&authorWithBooks)
    }

    return authorWithBooks, nil
}

// get all authors and return
func getAllAuthors() []model.AuthorWithBooks {
    var authors []model.AuthorWithBooks

    pipeline := []bson.M{
        {"$lookup": bson.M{
            "from":         "bookAuthor",
            "localField":   "_id",
            "foreignField": "author",
            "as":           "authorBookRelations",
        }},
        {"$lookup": bson.M{
            "from":         "bookList",
            "localField":   "authorBookRelations.book",
            "foreignField": "_id",
            "as":           "books",
        }},
        {"$project": bson.M{
            "_id":   1,
            "name":  1,
            "books": bson.M{"$ifNull": []interface{}{
                bson.M{"$map": bson.M{
                    "input": "$books",
                    "as":    "book",
                    "in": bson.M{"title": "$$book.title"},
                }},
                []model.BookInfo{},
            }},
        }},
    }

    cursor, err := collection.Aggregate(context.Background(), pipeline)

    if err != nil {
        log.Fatal(err)
    }

    for cursor.Next(context.Background()) {
        var author model.AuthorWithBooks
        cursor.Decode(&author)
        authors = append(authors, author)
    }

    fmt.Println("Found all documents: ", authors)

    return authors
}

func GetAllAuthors(c *gin.Context) {
    allAuthors := getAllAuthors()
    c.JSON(http.StatusOK, allAuthors)
}

func GetAuthor(c *gin.Context) {
    authorId := c.Param("authorId")
    author, err := getAuthor(authorId)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Author not found"})
        return
    }
    c.JSON(http.StatusOK, author)
}

func CreateAuthor(c *gin.Context) {
    var author model.Author
    if err := c.ShouldBindJSON(&author); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    author.Books = []primitive.ObjectID{} 
    insertAuthor(author)
    c.JSON(http.StatusOK, author)
}

func UpdateAuthor(c *gin.Context) {
    c.Writer.Header().Set("Content-Type", "application/json")
    c.Writer.Header().Set("Allow-Control-Allow-Methods", "PUT")
    authorId := c.Param("authorId")
    var author model.Author
    if err := c.ShouldBindJSON(&author); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    updateAuthor(authorId, author)
    c.JSON(http.StatusOK, gin.H{"status": "Updated"})
}

func DeleteAuthor(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Allow-Control-Allow-Methods", "DELETE")
	authorId := c.Param("authorId")
	deleteAuthor(authorId)
	c.JSON(http.StatusOK, gin.H{"status": "Deleted"})
}
