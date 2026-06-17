package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	PaymentStatusPending   = "pending"
	PaymentStatusApproved  = "approved"
	PaymentStatusRejected  = "rejected"
	PaymentStatusCancelled = "cancelled"
)

// Payment represents a payment attempt for a subscription plan
type Payment struct {
	ID             primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID         primitive.ObjectID `json:"userId" bson:"user_id"`
	RestaurantID   primitive.ObjectID `json:"restaurantId" bson:"restaurant_id"`
	PlanID         primitive.ObjectID `json:"planId" bson:"plan_id"`
	MPPreferenceID string             `json:"mpPreferenceId" bson:"mp_preference_id"`
	MPPaymentID    string             `json:"mpPaymentId,omitempty" bson:"mp_payment_id,omitempty"`
	Amount         float64            `json:"amount" bson:"amount"`
	Status         string             `json:"status" bson:"status"`
	CreatedAt      time.Time          `json:"createdAt" bson:"created_at"`
	UpdatedAt      time.Time          `json:"updatedAt" bson:"updated_at"`
}

// CreatePreferenceInput represents the request to initiate a payment
type CreatePreferenceInput struct {
	PlanID       string `json:"planId" binding:"required"`
	RestaurantID string `json:"restaurantId" binding:"required"`
}

// WebhookBody represents the payload MercadoPago sends to the webhook endpoint
type WebhookBody struct {
	Action string          `json:"action"`
	Data   WebhookBodyData `json:"data"`
}

type WebhookBodyData struct {
	ID string `json:"id"`
}

// PaymentEvent is the message sent over WebSocket when a payment status changes
type PaymentEvent struct {
	Type    string   `json:"type"`
	Payment *Payment `json:"payment"`
}

func NewPaymentApprovedEvent(payment *Payment) *PaymentEvent {
	return &PaymentEvent{
		Type:    "payment.approved",
		Payment: payment,
	}
}

func NewPayment(userID, restaurantID, planID primitive.ObjectID, mpPreferenceID string, amount float64) *Payment {
	now := time.Now()
	return &Payment{
		UserID:         userID,
		RestaurantID:   restaurantID,
		PlanID:         planID,
		MPPreferenceID: mpPreferenceID,
		Amount:         amount,
		Status:         PaymentStatusPending,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}
