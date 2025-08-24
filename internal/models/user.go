package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Email     string             `bson:"email"        json:"email"`
	Phone     string             `bson:"phone"        json:"phone"`
	Password  string             `bson:"password"     json:"-"`
	CreatedAt time.Time          `bson:"createdAt"    json:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt"    json:"updatedAt"`
}
