package models

import "time"

type User struct {
	ID        string    `bson:"_id,omitempty" json:"id"`
	Email     string    `bson:"email"        json:"email"`
	Phone     string    `bson:"phone"        json:"phone"`
	Password  string    `bson:"password"     json:"-"`
	CreatedAt time.Time `bson:"createdAt"    json:"createdAt"`
	UpdatedAt time.Time `bson:"updatedAt"    json:"updatedAt"`
}
