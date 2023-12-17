package model

import (
    "go.mongodb.org/mongo-driver/bson/primitive"
)

type BookAuthor struct {
    ID     primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
    Author primitive.ObjectID `json:"author,omitempty" bson:"author,omitempty"`
    Book   primitive.ObjectID `json:"book,omitempty" bson:"book,omitempty"`
}
