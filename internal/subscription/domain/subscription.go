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
	PlanNameBasico      = "Básico"
	PlanNameIntermedio  = "Intermedio"
	PlanNameProfesional = "Profesional"
)

// New plan names (business model redesign)
const (
	PlanNameStarter  = "Starter"
	PlanNamePro      = "Pro"
	PlanNameBusiness = "Business"
)

// Reports tiers per plan
const (
	ReportsTierNone         = "none"
	ReportsTierIncluded     = "included"
	ReportsTierAdvanced     = "advanced"
	ReportsTierConsolidated = "consolidated"
)

// Billing cycles
const (
	BillingCycleMonthly = "monthly"
	BillingCycleAnnual  = "annual"
)

// Unlimited is a sentinel value for MaxTables/MaxBranches meaning no limit
const Unlimited = -1

// Plan represents a subscription plan
type Plan struct {
	ID                  primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name                string             `json:"name" bson:"name"`
	Price               float64            `json:"price" bson:"price"`                               // ARS mensual (fuente de verdad para el cobro)
	PriceAnnual         float64            `json:"priceAnnual" bson:"price_annual"`                  // ARS anual
	PriceUSD            float64            `json:"priceUsd" bson:"price_usd"`                        // ancla USD mensual (referencia)
	ExtraBranchPrice    float64            `json:"extraBranchPrice" bson:"extra_branch_price"`       // ARS por sucursal extra (mensual); 0 = sin add-on
	ExtraBranchPriceUSD float64            `json:"extraBranchPriceUsd" bson:"extra_branch_price_usd"` // ancla USD por sucursal extra
	MaxTables           int                `json:"maxTables" bson:"max_tables"`
	IncludedBranches    int                `json:"includedBranches" bson:"included_branches"` // sucursales incluidas en la base
	MaxBranches         int                `json:"maxBranches" bson:"max_branches"`           // tope del tier (Unlimited = sin tope)
	TrialDays           int                `json:"trialDays" bson:"trial_days"`
	ReportsTier         string             `json:"reportsTier" bson:"reports_tier"`
	CreatedAt           time.Time          `json:"createdAt" bson:"created_at"`
	UpdatedAt           time.Time          `json:"updatedAt" bson:"updated_at"`
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
	PurchasedBranches     int                `json:"purchasedBranches" bson:"purchased_branches"`
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
func NewPlan(name string, price, priceAnnual, priceUSD, extraBranchPrice, extraBranchPriceUSD float64, maxTables, includedBranches, maxBranches, trialDays int, reportsTier string) *Plan {
	now := time.Now()
	return &Plan{
		Name:                name,
		Price:               price,
		PriceAnnual:         priceAnnual,
		PriceUSD:            priceUSD,
		ExtraBranchPrice:    extraBranchPrice,
		ExtraBranchPriceUSD: extraBranchPriceUSD,
		MaxTables:           maxTables,
		IncludedBranches:    includedBranches,
		MaxBranches:         maxBranches,
		TrialDays:           trialDays,
		ReportsTier:         reportsTier,
		CreatedAt:           now,
		UpdatedAt:           now,
	}
}

// PriceForCycle returns the ARS amount to charge for the given billing cycle.
// Anything other than the annual cycle falls back to the monthly price.
func (p *Plan) PriceForCycle(cycle string) float64 {
	if cycle == BillingCycleAnnual {
		return p.PriceAnnual
	}
	return p.Price
}

// AnnualCycleMultiplier: el ciclo anual cobra 10 meses (2 gratis).
const AnnualCycleMultiplier = 10

// SupportsBranchAddon indica si el plan permite sumar sucursales extra.
func (p *Plan) SupportsBranchAddon() bool {
	return p.ExtraBranchPriceUSD > 0
}

// PriceForBranches devuelve el monto ARS a cobrar para el ciclo y la cantidad
// de sucursales elegida. Suma el precio de cada sucursal por encima de las
// incluidas; en anual, base y extras aplican el multiplicador (2 meses gratis).
func (p *Plan) PriceForBranches(cycle string, branches int) float64 {
	extras := branches - p.IncludedBranches
	if extras < 0 {
		extras = 0
	}
	if cycle == BillingCycleAnnual {
		return p.PriceAnnual + float64(extras)*p.ExtraBranchPrice*AnnualCycleMultiplier
	}
	return p.Price + float64(extras)*p.ExtraBranchPrice
}

// SubscriptionWithPlan represents a subscription with its associated plan embedded
type SubscriptionWithPlan struct {
	ID                    primitive.ObjectID `json:"id"`
	UserID                primitive.ObjectID `json:"userId"`
	RestaurantID          primitive.ObjectID `json:"restaurantId"`
	PlanID                primitive.ObjectID `json:"planId"`
	PurchasedBranches     int                `json:"purchasedBranches"`
	Plan                  *Plan              `json:"plan"`
	Status                string             `json:"status"`
	TrialStartedAt        *time.Time         `json:"trialStartedAt,omitempty"`
	TrialEndsAt           *time.Time         `json:"trialEndsAt,omitempty"`
	PaymentSubscriptionID string             `json:"paymentSubscriptionId,omitempty"`
	CreatedAt             time.Time          `json:"createdAt"`
	UpdatedAt             time.Time          `json:"updatedAt"`
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
