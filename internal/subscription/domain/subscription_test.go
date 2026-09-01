package domain

import (
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestNewPlanSetsAllFields(t *testing.T) {
	p := NewPlan(PlanNameStarter, 28500, 285000, 19, 0, 0, 25, 1, 1, 30, ReportsTierNone)
	if p.Name != "Starter" {
		t.Fatalf("name = %q, want Starter", p.Name)
	}
	if p.Price != 28500 || p.PriceAnnual != 285000 || p.PriceUSD != 19 {
		t.Fatalf("prices = %v/%v/%v", p.Price, p.PriceAnnual, p.PriceUSD)
	}
	if p.MaxTables != 25 || p.IncludedBranches != 1 || p.MaxBranches != 1 || p.TrialDays != 30 {
		t.Fatalf("limits = %v/%v/%v/%v", p.MaxTables, p.IncludedBranches, p.MaxBranches, p.TrialDays)
	}
	if p.ReportsTier != ReportsTierNone {
		t.Fatalf("reports=%q", p.ReportsTier)
	}
	if p.SupportsBranchAddon() {
		t.Fatalf("Starter no debería soportar add-on de sucursales")
	}
}

func TestPriceForCycle(t *testing.T) {
	p := NewPlan(PlanNamePro, 73500, 735000, 49, 0, 0, 40, 3, 3, 30, ReportsTierIncluded)
	if got := p.PriceForCycle(BillingCycleMonthly); got != 73500 {
		t.Fatalf("monthly = %v, want 73500", got)
	}
	if got := p.PriceForCycle(BillingCycleAnnual); got != 735000 {
		t.Fatalf("annual = %v, want 735000", got)
	}
	if got := p.PriceForCycle(""); got != 73500 {
		t.Fatalf("empty cycle = %v, want monthly 73500", got)
	}
}

func TestPriceForBranchesBusiness(t *testing.T) {
	// Business: base 118500 (anual 1185000), extra 13500 ARS, incluye 5.
	b := NewPlan(PlanNameBusiness, 118500, 1185000, 79, 13500, 9, Unlimited, 5, Unlimited, 30, ReportsTierAdvanced)
	if !b.SupportsBranchAddon() {
		t.Fatalf("Business debería soportar add-on")
	}
	if got := b.PriceForBranches(BillingCycleMonthly, 5); got != 118500 {
		t.Fatalf("5 suc mensual = %v, want 118500", got)
	}
	if got := b.PriceForBranches(BillingCycleMonthly, 8); got != 159000 { // 118500 + 3*13500
		t.Fatalf("8 suc mensual = %v, want 159000", got)
	}
	if got := b.PriceForBranches(BillingCycleMonthly, 3); got != 118500 { // por debajo del incluido no resta
		t.Fatalf("3 suc mensual = %v, want 118500", got)
	}
	if got := b.PriceForBranches(BillingCycleAnnual, 8); got != 1590000 { // 1185000 + 3*13500*10
		t.Fatalf("8 suc anual = %v, want 1590000", got)
	}
}

func TestPriceForBranchesNoAddon(t *testing.T) {
	// Un plan sin add-on ignora las sucursales extra (extra price = 0).
	s := NewPlan(PlanNameStarter, 28500, 285000, 19, 0, 0, 25, 1, 1, 30, ReportsTierNone)
	if got := s.PriceForBranches(BillingCycleMonthly, 5); got != 28500 {
		t.Fatalf("starter 5 suc = %v, want 28500", got)
	}
}

func TestApplyCompedSetsActiveBusinessAndClearsTrial(t *testing.T) {
	businessID := primitive.NewObjectID()
	now := time.Now()
	// Suscripción vencida en trial: la cuenta cortesía debe recuperarla.
	sub := &Subscription{
		Status:            SubscriptionStatusExpired,
		PlanID:            primitive.NewObjectID(),
		PurchasedBranches: 1,
		TrialStartedAt:    &now,
		TrialEndsAt:       &now,
	}

	sub.ApplyComped(businessID)

	if sub.Status != SubscriptionStatusActive {
		t.Fatalf("status = %q, want active", sub.Status)
	}
	if sub.PlanID != businessID {
		t.Fatalf("planID = %v, want %v", sub.PlanID, businessID)
	}
	if sub.PurchasedBranches != CompedBranches {
		t.Fatalf("purchasedBranches = %d, want %d", sub.PurchasedBranches, CompedBranches)
	}
	if sub.TrialStartedAt != nil || sub.TrialEndsAt != nil {
		t.Fatalf("trial no fue limpiado: started=%v ends=%v", sub.TrialStartedAt, sub.TrialEndsAt)
	}
	if CompedBranches <= 0 {
		t.Fatalf("CompedBranches debe ser > 0 (el chequeo de límite rechaza <= 0)")
	}
}

func TestIsCompedActiveIdempotency(t *testing.T) {
	businessID := primitive.NewObjectID()

	active := &Subscription{Status: SubscriptionStatusActive, PlanID: businessID, PurchasedBranches: CompedBranches}
	if !active.IsCompedActive(businessID) {
		t.Fatalf("una suscripción ya cortesía debería reportar IsCompedActive=true")
	}

	// Distinto plan, estado no-activo, o pocas sucursales => no está lista.
	cases := []*Subscription{
		{Status: SubscriptionStatusActive, PlanID: primitive.NewObjectID(), PurchasedBranches: CompedBranches},
		{Status: SubscriptionStatusExpired, PlanID: businessID, PurchasedBranches: CompedBranches},
		{Status: SubscriptionStatusActive, PlanID: businessID, PurchasedBranches: 1},
	}
	for i, c := range cases {
		if c.IsCompedActive(businessID) {
			t.Fatalf("caso %d: IsCompedActive debería ser false", i)
		}
	}
}
