package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Subscription statuses
const (
	SubscriptionStatusActive   = "active"
	SubscriptionStatusTrialing = "trialing"
	SubscriptionStatusCanceled = "canceled"
	SubscriptionStatusExpired  = "expired"
	SubscriptionStatusPastDue  = "past_due"
)

// Plan names
const (
	PlanNameInicial     = "Inicial"
	PlanNameProfesional = "Profesional"
)

// Unlimited is a sentinel value for MaxTables/MaxBranches meaning no limit
const Unlimited = -1

// Plan represents a subscription plan
type Plan struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name        string             `json:"name" bson:"name"`
	Price       float64            `json:"price" bson:"price"`
	MaxTables   int                `json:"maxTables" bson:"max_tables"`
	MaxBranches int                `json:"maxBranches" bson:"max_branches"`
	TrialDays   int                `json:"trialDays" bson:"trial_days"`
	CreatedAt   time.Time          `json:"createdAt" bson:"created_at"`
	UpdatedAt   time.Time          `json:"updatedAt" bson:"updated_at"`
}

// Subscription represents a user's subscription to a plan
type Subscription struct {
	ID                    primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID                primitive.ObjectID `json:"userId" bson:"user_id"`
	RestaurantID          primitive.ObjectID `json:"restaurantId" bson:"restaurant_id"`
	PlanID                primitive.ObjectID `json:"planId" bson:"plan_id"`
	Status                string             `json:"status" bson:"status"`
	TrialStartedAt        *time.Time         `json:"trialStartedAt,omitempty" bson:"trial_started_at,omitempty"`
	TrialEndsAt           *time.Time         `json:"trialEndsAt,omitempty" bson:"trial_ends_at,omitempty"`
	PaymentSubscriptionID string             `json:"paymentSubscriptionId,omitempty" bson:"payment_subscription_id,omitempty"`
	CreatedAt             time.Time          `json:"createdAt" bson:"created_at"`
	UpdatedAt             time.Time          `json:"updatedAt" bson:"updated_at"`
}

// CreateSubscriptionInput represents the data needed to create a subscription
type CreateSubscriptionInput struct {
	RestaurantID          string `json:"restaurantId" binding:"required"`
	PlanID                string `json:"planId" binding:"required"`
	StartTrial            bool   `json:"startTrial"`
	PaymentSubscriptionID string `json:"paymentSubscriptionId,omitempty"`
}

// UpdateSubscriptionInput represents the data needed to update a subscription
type UpdateSubscriptionInput struct {
	PlanID                string `json:"planId,omitempty"`
	Status                string `json:"status,omitempty" binding:"omitempty,oneof=active trialing canceled expired past_due"`
	PaymentSubscriptionID string `json:"paymentSubscriptionId,omitempty"`
}

// NewPlan creates a new plan with the current timestamp
func NewPlan(name string, price float64, maxTables, maxBranches, trialDays int) *Plan {
	now := time.Now()
	return &Plan{
		Name:        name,
		Price:       price,
		MaxTables:   maxTables,
		MaxBranches: maxBranches,
		TrialDays:   trialDays,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// NewSubscription creates a new subscription with the current timestamp
func NewSubscription(userID, restaurantID, planID primitive.ObjectID, status string) *Subscription {
	now := time.Now()
	return &Subscription{
		UserID:       userID,
		RestaurantID: restaurantID,
		PlanID:       planID,
		Status:       status,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}
