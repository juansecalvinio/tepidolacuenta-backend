package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Table struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	BranchID  primitive.ObjectID `json:"branchId" bson:"branch_id"`
	Number    int                `json:"number" bson:"number"`
	QRCode    string             `json:"qrCode" bson:"qr_code"`
	IsActive  bool               `json:"isActive" bson:"is_active"`
	CreatedAt time.Time          `json:"createdAt" bson:"created_at"`
	UpdatedAt time.Time          `json:"updatedAt" bson:"updated_at"`
}

// CreateTableInput represents the data needed to create a table
type CreateTableInput struct {
	BranchID string `json:"branchId" binding:"required"`
	Number   int    `json:"number" binding:"required,min=1"`
}

// UpdateTableInput represents the data needed to update a table
type UpdateTableInput struct {
	Number   int   `json:"number,omitempty" binding:"omitempty,min=1"`
	IsActive *bool `json:"isActive,omitempty"`
}

// BulkCreateTablesInput represents the data needed to create multiple tables
type BulkCreateTablesInput struct {
	BranchID string `json:"branchId" binding:"required"`
	Count    int    `json:"count" binding:"required,min=1,max=100"`
}

// NewTable creates a new table with the current timestamp
func NewTable(branchID primitive.ObjectID, number int, qrCode string) *Table {
	now := time.Now()
	return &Table{
		BranchID:  branchID,
		Number:    number,
		QRCode:    qrCode,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
