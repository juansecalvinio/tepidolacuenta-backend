package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Table struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	RestaurantID primitive.ObjectID `json:"restaurantId" bson:"restaurant_id"`
	Number       int                `json:"number" bson:"number"`
	Capacity     int                `json:"capacity" bson:"capacity"`
	QRCode       string             `json:"qrCode" bson:"qr_code"`
	IsActive     bool               `json:"isActive" bson:"is_active"`
	CreatedAt    time.Time          `json:"createdAt" bson:"created_at"`
	UpdatedAt    time.Time          `json:"updatedAt" bson:"updated_at"`
}

// CreateTableInput represents the data needed to create a table
type CreateTableInput struct {
	RestaurantID string `json:"restaurantId" binding:"required"`
	Number       int    `json:"number" binding:"required,min=1"`
	Capacity     int    `json:"capacity" binding:"required,min=1,max=20"`
}

// UpdateTableInput represents the data needed to update a table
type UpdateTableInput struct {
	Number   int  `json:"number,omitempty" binding:"omitempty,min=1"`
	Capacity int  `json:"capacity,omitempty" binding:"omitempty,min=1,max=20"`
	IsActive *bool `json:"isActive,omitempty"`
}

// NewTable creates a new table with the current timestamp
func NewTable(restaurantID primitive.ObjectID, number, capacity int, qrCode string) *Table {
	now := time.Now()
	return &Table{
		RestaurantID: restaurantID,
		Number:       number,
		Capacity:     capacity,
		QRCode:       qrCode,
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}
