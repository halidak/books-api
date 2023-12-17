package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Book struct {
    ID     primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
    Title  string             `json:"title,omitempty" bson:"title,omitempty"`
    Genre  string             `json:"genre,omitempty" bson:"genre,omitempty"`
    Authors []primitive.ObjectID `json:"authors,omitempty" bson:"authors,omitempty"`
    Read   bool               `json:"read,omitempty"`
}