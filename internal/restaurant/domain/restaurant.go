package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Restaurant struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID      primitive.ObjectID `json:"userId" bson:"user_id"`
	Name        string             `json:"name" bson:"name"`
	Address     string             `json:"address" bson:"address"`
	Phone       string             `json:"phone" bson:"phone"`
	Description string             `json:"description,omitempty" bson:"description,omitempty"`
	CreatedAt   time.Time          `json:"createdAt" bson:"created_at"`
	UpdatedAt   time.Time          `json:"updatedAt" bson:"updated_at"`
}

// CreateRestaurantInput represents the data needed to create a restaurant
type CreateRestaurantInput struct {
	Name        string `json:"name" binding:"required,min=3,max=100"`
	Address     string `json:"address" binding:"required,min=5,max=200"`
	Phone       string `json:"phone" binding:"required,min=8,max=20"`
	Description string `json:"description,omitempty" binding:"max=500"`
}

// UpdateRestaurantInput represents the data needed to update a restaurant
type UpdateRestaurantInput struct {
	Name        string `json:"name,omitempty" binding:"omitempty,min=3,max=100"`
	Address     string `json:"address,omitempty" binding:"omitempty,min=5,max=200"`
	Phone       string `json:"phone,omitempty" binding:"omitempty,min=8,max=20"`
	Description string `json:"description,omitempty" binding:"max=500"`
}

// NewRestaurant creates a new restaurant with the current timestamp
func NewRestaurant(userID primitive.ObjectID, name, address, phone, description string) *Restaurant {
	now := time.Now()
	return &Restaurant{
		UserID:      userID,
		Name:        name,
		Address:     address,
		Phone:       phone,
		Description: description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}
