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
var bookAuthor *mongo.Collection

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
	colName3 := os.Getenv("COLNAME3")

	bookCollection = client.Database(dbName).Collection(colName2)
	readingListCollection = client.Database(dbName).Collection(colName)
	bookAuthor = client.Database(dbName).Collection(colName3)

	fmt.Println("Collection istance is ready")
}

// insert book with author
func insertBook(book model.Book, authorIDs []primitive.ObjectID) {
    inserted, err := bookCollection.InsertOne(context.Background(), book)

    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Inserted a single document: ", inserted.InsertedID)

    for _, authorID := range authorIDs {
        _, err := readingListCollection.UpdateOne(
            context.Background(),
            bson.M{"_id": authorID},
            bson.M{"$addToSet": bson.M{"books": inserted.InsertedID}},
        )
        if err != nil {
            log.Fatal(err)
        }
    }

    for _, authorID := range authorIDs {
        _, err := bookAuthor.InsertOne(context.Background(), bson.M{"book": inserted.InsertedID, "author": authorID})
        if err != nil {
            log.Fatal(err)
        }
    }
}

// get book with author name
func getBookWithAuthor(bookId string) (model.BookWithAuthor, error) {
	var bookWithAuthor model.BookWithAuthor

	id, err := primitive.ObjectIDFromHex(bookId)
	if err != nil {
		return bookWithAuthor, err
	}

    pipeline := []bson.M{
        {"$match": bson.M{"_id": id}},
        {"$lookup": bson.M{
            "from":         "bookAuthor",
            "localField":   "_id",
            "foreignField": "book",
            "as":           "bookAuthorRelations",
        }},
        {"$lookup": bson.M{
            "from":         "readList",
            "localField":   "bookAuthorRelations.author",
            "foreignField": "_id",
            "as":           "authors",
        }},
        {"$project": bson.M{
            "_id":     1,
            "title":   1,
            "genre":   1,
            "authors": bson.M{"$ifNull": []interface{}{
                bson.M{"$map": bson.M{
                    "input": "$authors",
                    "as":    "author",
                    "in": bson.M{"name": "$$author.name"},
                }},
                []model.AuthorInfo{},
            }},
            "read":    1,
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

// get all book with author name
func getAllBooksWithAuthors() []model.BookWithAuthor {
	var booksWithAuthors []model.BookWithAuthor

    pipeline := []bson.M{
        {"$lookup": bson.M{
            "from":         "bookAuthor",
            "localField":   "_id",
            "foreignField": "book",
            "as":           "bookAuthorRelations",
        }},
        {"$lookup": bson.M{
            "from":         "readList",
            "localField":   "bookAuthorRelations.author",
            "foreignField": "_id",
            "as":           "authorInfo",
        }},
        {"$unwind": "$authorInfo"},
        {"$group": bson.M{
            "_id": "$_id",
            "title": bson.M{"$first": "$title"},
            "genre": bson.M{"$first": "$genre"},
            "read": bson.M{"$first": "$read"},
            "authors": bson.M{"$push": bson.M{"_id": "$authorInfo._id", "name": "$authorInfo.name"}},
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

// update book
func updateBook(bookID string, book model.Book, authors []primitive.ObjectID) {
    id, _ := primitive.ObjectIDFromHex(bookID)
    filter := bson.M{"_id": id}
    update := bson.M{"$set": bson.M{"title": book.Title, "genre": book.Genre, "read": book.Read}}

    result, err := bookCollection.UpdateOne(context.Background(), filter, update)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Updated a single document: ", result.UpsertedID)

    // Update authors
    authorFilter := bson.M{"book": id}
    authorUpdate := bson.M{"$set": bson.M{"author": authors}}

    authorResult, err := bookAuthor.UpdateOne(context.Background(), authorFilter, authorUpdate)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Updated authors: ", authorResult.ModifiedCount)
}


// delete book
func deleteBook(bookId string) {
    id, _ := primitive.ObjectIDFromHex(bookId)
    filter := bson.M{"_id": id}

    // Delete the book from the books collection
    result, err := bookCollection.DeleteOne(context.Background(), filter)

    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Deleted a single document: ", result.DeletedCount)

    // Delete the book from the bookAuthor collection
    authorFilter := bson.M{"book": id}
    authorResult, err := bookAuthor.DeleteMany(context.Background(), authorFilter)

    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Deleted from bookAuthor: ", authorResult.DeletedCount)

    // Delete the book from the readingList collection
    readListFilter := bson.M{"books": bson.M{"$in": []primitive.ObjectID{id}}}
    update := bson.M{"$pull": bson.M{"books": id}}
    readListResult, err := readingListCollection.UpdateMany(context.Background(), readListFilter, update)

    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Deleted from readingList: ", readListResult.ModifiedCount)
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

    authorIDs := make([]primitive.ObjectID, len(book.Authors))
    for i, strID := range book.Authors {
        authorID, err := primitive.ObjectIDFromHex(strID.Hex())
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid author ID"})
            return
        }
        authorIDs[i] = authorID
    }

    if exist, err := authorsExist(authorIDs); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking author existence"})
        return
    } else if !exist {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Some authors do not exist"})
        return
    }

    insertBook(book, authorIDs)

    c.JSON(http.StatusOK, book)
}

func authorsExist(authorIDs []primitive.ObjectID) (bool, error) {
    filter := bson.M{"_id": bson.M{"$in": authorIDs}}

    count, err := readingListCollection.CountDocuments(context.Background(), filter)
    if err != nil {
        return false, err
    }

    return count == int64(len(authorIDs)), nil
}

func UpdateBook(c *gin.Context) {
    bookId := c.Param("bookId")
    var book model.Book
    if err := c.ShouldBindJSON(&book); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    authorIDs := make([]primitive.ObjectID, len(book.Authors))
    for i, strID := range book.Authors {
        authorID, err := primitive.ObjectIDFromHex(strID.Hex())
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid author ID"})
            return
        }
        authorIDs[i] = authorID
    }

    // Check if authors exist in the database
    if exist, err := authorsExist(authorIDs); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking author existence"})
        return
    } else if !exist {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Some authors do not exist"})
        return
    }

    updateBook(bookId, book, authorIDs)

    c.JSON(http.StatusOK, book)
}

func DeleteBook(c *gin.Context) {
	bookId := c.Param("bookId")
	deleteBook(bookId)
	c.JSON(http.StatusOK, gin.H{"message": "Book deleted"})
}

// get all books from author
func getAllBooksForAuthor(authorId primitive.ObjectID) []model.Book {
    var bookAuthorLinks []model.BookAuthor
    var books []model.Book

    linkFilter := bson.M{"author": authorId}
    cursor, err := bookAuthor.Find(context.Background(), linkFilter)
    if err != nil {
        log.Fatal(err)
    }

    for cursor.Next(context.Background()) {
        var link model.BookAuthor
        cursor.Decode(&link)
        bookAuthorLinks = append(bookAuthorLinks, link)
    }

    for _, link := range bookAuthorLinks {
        bookFilter := bson.M{"_id": link.Book}
        var book model.Book
        err := bookCollection.FindOne(context.Background(), bookFilter).Decode(&book)
        if err != nil {
            log.Fatal(err)
        }
        books = append(books, book)
    }

    return books
}

// get all books from author
func GetBooksForAuthor(c *gin.Context) {
	authorId := c.Param("authorId")

	objAuthorId, err := primitive.ObjectIDFromHex(authorId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid author ID"})
		return
	}

	booksForAuthor := getAllBooksForAuthor(objAuthorId)

	c.JSON(http.StatusOK, booksForAuthor)
}

// read book
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
