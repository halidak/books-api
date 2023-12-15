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

	collection = client.Database(dbName).Collection(colName)

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

	result, err := collection.DeleteOne(context.Background(), filter)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Deleted a single document: ", result.DeletedCount)
}

// get author and return
func getAuthor(authorId string) (model.Author, error) {
    var author model.Author

    id, err := primitive.ObjectIDFromHex(authorId)
    if err != nil {
        return author, err
    }
    filter := bson.M{"_id": id}

    err = collection.FindOne(context.Background(), filter).Decode(&author)
    if err != nil {
        return author, err
    }

    fmt.Println("Found a single document: ", author)

    return author, nil
}
// get all authors and return
func getAllAuthors() []model.Author {
	var authors []model.Author

	cursor, err := collection.Find(context.Background(), bson.M{})

	if err != nil {
		log.Fatal(err)
	}

	for cursor.Next(context.Background()) {
		var author model.Author
		cursor.Decode(&author)
		authors = append(authors, author)
	}

	fmt.Println("Found a all documents: ", authors)

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
