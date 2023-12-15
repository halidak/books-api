package model

import (
    "go.mongodb.org/mongo-driver/bson/primitive"
)

type Author struct {
    ID    primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
    Name  string             `json:"name,omitempty" bson:"name,omitempty"`
    Books []primitive.ObjectID `json:"books,omitempty" bson:"books,omitempty"`
}