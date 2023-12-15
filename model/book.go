package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Book struct {
    ID     primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
    Title  string             `json:"title,omitempty" bson:"title,omitempty"`
    Author primitive.ObjectID `json:"author,omitempty" bson:"author,omitempty"`
    Read   bool               `json:"read,omitempty"`
}