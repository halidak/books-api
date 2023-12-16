package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type AuthorWithBooks struct {
    ID    primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
    Name  string             `json:"name,omitempty" bson:"name,omitempty"`
    Books []BookInfo         `json:"books,omitempty" bson:"books,omitempty"`
}

type BookInfo struct {
    Title string             `json:"title,omitempty" bson:"title,omitempty"`
}