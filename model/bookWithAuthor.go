package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type BookWithAuthor struct {
    ID     primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
    Title  string             `json:"title,omitempty" bson:"title,omitempty"`
    Genre  string             `json:"genre,omitempty" bson:"genre,omitempty"`
    Authors []AuthorInfo      `json:"authors,omitempty" bson:"authors,omitempty"`
    Read   bool               `json:"read,omitempty" bson:"read,omitempty"`
}

type AuthorInfo struct {
	Name string             `json:"name,omitempty" bson:"name,omitempty"`
}
