package migration

import (
	"context"
	"time"

	"juansecalvinio/tepidolacuenta/internal/subscription/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// All returns the ordered list of migrations to apply
func All() []Migration {
	return []Migration{
		{
			Name: "001_update_plan_values",
			Run:  updatePlanValues,
		},
		{
			Name: "002_clean_collections",
			Run:  cleanCollections,
		},
		{
			Name: "003_rename_plan_inicial_to_basico",
			Run:  renamePlanInicialToBasico,
		},
		{
			Name: "004_update_plan_values",
			Run:  updatePlanValues,
		},
		{
			Name: "005_update_plan_values",
			Run:  updatePlanValues,
		},
		{
			Name: "006_clean_collections",
			Run:  cleanCollections,
		},
		{
			Name: "007_clean_collections",
			Run:  cleanCollections,
		},
		{
			Name: "008_add_role_owner_to_existing_users",
			Run:  addRoleOwnerToExistingUsers,
		},
		{
			Name: "009_create_invitations_index",
			Run:  createInvitationsIndex,
		},
		{
			Name: "010_update_plan_prices",
			Run:  updatePlanPrices,
		},
		{
			Name: "011_update_plan_prices",
			Run:  updatePlanPrices,
		},
		{
			Name: "012_reset_all_data",
			Run:  resetAllData,
		},
	}
}

// updatePlanPrices updates only the price field of existing plans
func updatePlanPrices(ctx context.Context, db *mongo.Database) error {
	plans := db.Collection("plans")

	prices := map[string]int{
		domain.PlanNameBasico:      19999,
		domain.PlanNameIntermedio:  49999,
		domain.PlanNameProfesional: 99999,
	}

	for name, price := range prices {
		_, err := plans.UpdateOne(
			ctx,
			bson.M{"name": name},
			bson.M{"$set": bson.M{
				"price":      price,
				"updated_at": time.Now(),
			}},
		)
		if err != nil && err != mongo.ErrNoDocuments {
			return err
		}
	}

	return nil
}

// updatePlanValues sets the correct prices, limits and trial days for both plans
func updatePlanValues(ctx context.Context, db *mongo.Database) error {
	plans := db.Collection("plans")

	updates := []struct {
		name   string
		fields bson.M
	}{
		{
			name: domain.PlanNameBasico,
			fields: bson.M{
				"price":        19999,
				"max_tables":   20,
				"max_branches": 1,
				"trial_days":   30,
				"updated_at":   time.Now(),
			},
		},
		{
			name: domain.PlanNameIntermedio,
			fields: bson.M{
				"price":        49999,
				"max_tables":   50,
				"max_branches": 3,
				"trial_days":   30,
				"updated_at":   time.Now(),
			},
		},
		{
			name: domain.PlanNameProfesional,
			fields: bson.M{
				"price":        99999,
				"max_tables":   domain.Unlimited,
				"max_branches": domain.Unlimited,
				"trial_days":   30,
				"updated_at":   time.Now(),
			},
		},
	}

	for _, u := range updates {
		_, err := plans.UpdateOne(
			ctx,
			bson.M{"name": u.name},
			bson.M{"$set": u.fields},
			// upsert:false — solo actualiza si el plan ya existe (sembrado por seedPlans)
		)
		if err != nil && err != mongo.ErrNoDocuments {
			return err
		}
	}

	return nil
}

// renamePlanInicialToBasico renames the plan "Inicial" to "Básico" in the plans collection
func renamePlanInicialToBasico(ctx context.Context, db *mongo.Database) error {
	plans := db.Collection("plans")

	_, err := plans.UpdateOne(
		ctx,
		bson.M{"name": "Inicial"},
		bson.M{"$set": bson.M{
			"name":       domain.PlanNameBasico,
			"updated_at": time.Now(),
		}},
	)
	return err
}

// addRoleOwnerToExistingUsers sets role:"owner" on all users that don't have a role yet
func addRoleOwnerToExistingUsers(ctx context.Context, db *mongo.Database) error {
	users := db.Collection("users")
	_, err := users.UpdateMany(
		ctx,
		bson.M{"role": bson.M{"$exists": false}},
		bson.M{"$set": bson.M{"role": "owner"}},
	)
	return err
}

// createInvitationsIndex creates a unique index on invitations.code
func createInvitationsIndex(ctx context.Context, db *mongo.Database) error {
	invitations := db.Collection("invitations")
	_, err := invitations.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "code", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	return err
}

// cleanCollections drops the branches, users, restaurants, tables and requests collections
func cleanCollections(ctx context.Context, db *mongo.Database) error {
	collections := []string{"branches", "users", "restaurants", "tables", "requests", "plans"}
	for _, name := range collections {
		if err := db.Collection(name).Drop(ctx); err != nil {
			return err
		}
	}
	return nil
}

// resetAllData drops every data collection to leave the database clean for testing
// from scratch. It includes the newer collections (invitations, subscriptions,
// payments) that cleanCollections doesn't, so no orphan documents survive.
// "plans" is dropped too and re-seeded by seedPlans after migrations run.
// "migrations" is intentionally preserved — it's the runner's tracking table.
func resetAllData(ctx context.Context, db *mongo.Database) error {
	collections := []string{
		"branches",
		"users",
		"restaurants",
		"tables",
		"requests",
		"invitations",
		"subscriptions",
		"payments",
		"plans",
	}
	for _, name := range collections {
		if err := db.Collection(name).Drop(ctx); err != nil {
			return err
		}
	}
	return nil
}
