package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// RequestStatus represents the status of an account request
type RequestStatus string

const (
	StatusPending   RequestStatus = "pending"
	StatusAttended  RequestStatus = "attended"
	StatusCancelled RequestStatus = "cancelled"
)

// Request represents an account request from a table
type Request struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	RestaurantID primitive.ObjectID `bson:"restaurantId" json:"restaurantId"`
	BranchID     primitive.ObjectID `bson:"branchId" json:"branchId"`
	TableID      primitive.ObjectID `bson:"tableId" json:"tableId"`
	TableNumber  int                `bson:"tableNumber" json:"tableNumber"`
	Status       RequestStatus      `bson:"status" json:"status"`
	CreatedAt    time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt    time.Time          `bson:"updatedAt" json:"updatedAt"`
}

// CreateRequestInput represents the input for creating a request
type CreateRequestInput struct {
	RestaurantID string `json:"restaurantId" binding:"required"`
	BranchID     string `json:"branchId" binding:"required"`
	TableID      string `json:"tableId" binding:"required"`
	TableNumber  int    `json:"tableNumber" binding:"required,min=1"`
	Hash         string `json:"hash" binding:"required"`
}

// UpdateRequestStatusInput represents the input for updating request status
type UpdateRequestStatusInput struct {
	Status string `json:"status" binding:"required,oneof=pending attended cancelled"`
}

// NewRequest creates a new request
func NewRequest(restaurantID, branchID, tableID primitive.ObjectID, tableNumber int) *Request {
	now := time.Now()
	return &Request{
		ID:           primitive.NewObjectID(),
		RestaurantID: restaurantID,
		BranchID:     branchID,
		TableID:      tableID,
		TableNumber:  tableNumber,
		Status:       StatusPending,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// UpdateStatus updates the request status
func (r *Request) UpdateStatus(status RequestStatus) {
	r.Status = status
	r.UpdatedAt = time.Now()
}
