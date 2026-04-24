package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Invitation struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	RestaurantID primitive.ObjectID `json:"restaurantId" bson:"restaurant_id"`
	Code         string             `json:"code" bson:"code"`
	ExpiresAt    time.Time          `json:"expiresAt" bson:"expires_at"`
	Used         bool               `json:"used" bson:"used"`
	CreatedAt    time.Time          `json:"createdAt" bson:"created_at"`
}

type GenerateInvitationInput struct {
	RestaurantID string `json:"restaurantId" binding:"required"`
}

type GenerateInvitationResponse struct {
	Code      string    `json:"code"`
	ExpiresAt time.Time `json:"expiresAt"`
}

func NewInvitation(restaurantID primitive.ObjectID, code string) *Invitation {
	now := time.Now()
	return &Invitation{
		RestaurantID: restaurantID,
		Code:         code,
		ExpiresAt:    now.Add(7 * 24 * time.Hour),
		Used:         false,
		CreatedAt:    now,
	}
}
