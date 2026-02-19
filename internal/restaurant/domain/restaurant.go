package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Restaurant struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID    primitive.ObjectID `json:"userId" bson:"user_id"`
	Name      string             `json:"name" bson:"name"`
	CUIT      string             `json:"cuit" bson:"cuit,unique"`
	CreatedAt time.Time          `json:"createdAt" bson:"created_at"`
	UpdatedAt time.Time          `json:"updatedAt" bson:"updated_at"`
}

// CreateRestaurantInput represents the data needed to create a restaurant
type CreateRestaurantInput struct {
	Name string `json:"name" binding:"required,min=3,max=100"`
	CUIT string `json:"cuit" binding:"required"`
}

// UpdateRestaurantInput represents the data needed to update a restaurant
type UpdateRestaurantInput struct {
	Name string `json:"name,omitempty" binding:"omitempty,min=3,max=100"`
	CUIT string `json:"cuit,omitempty" binding:"omitempty"`
}

// NewRestaurant creates a new restaurant with the current timestamp
func NewRestaurant(userID primitive.ObjectID, name string, cuit string) *Restaurant {
	now := time.Now()
	return &Restaurant{
		UserID:    userID,
		Name:      name,
		CUIT:      cuit,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
