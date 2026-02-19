package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Branch struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	RestaurantID primitive.ObjectID `json:"restaurantId" bson:"restaurant_id"`
	Address      string             `json:"address" bson:"address"`
	Description  string             `json:"description,omitempty" bson:"description,omitempty"`
	IsActive     bool               `json:"isActive" bson:"is_active"`
	CreatedAt    time.Time          `json:"createdAt" bson:"created_at"`
	UpdatedAt    time.Time          `json:"updatedAt" bson:"updated_at"`
}

// CreateBranchInput represents the data needed to create a branch
type CreateBranchInput struct {
	RestaurantID string `json:"restaurantId" binding:"required"`
	Address      string `json:"address" binding:"required,max=200"`
	Description  string `json:"description,omitempty" binding:"max=500"`
}

// UpdateBranchInput represents the data needed to update a branch
type UpdateBranchInput struct {
	Address     string `json:"address,omitempty" binding:"omitempty,max=200"`
	Description string `json:"description,omitempty" binding:"max=500"`
	IsActive    *bool  `json:"isActive,omitempty"`
}

// NewBranch creates a new branch with the current timestamp
func NewBranch(restaurantID primitive.ObjectID, address, description string) *Branch {
	now := time.Now()
	return &Branch{
		RestaurantID: restaurantID,
		Address:      address,
		Description:  description,
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}
