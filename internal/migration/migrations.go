package migration

import (
	"context"
	"time"

	"juansecalvinio/tepidolacuenta/internal/subscription/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// All returns the ordered list of migrations to apply
func All() []Migration {
	return []Migration{
		{
			Name: "001_update_plan_values",
			Run:  updatePlanValues,
		},
	}
}

// updatePlanValues sets the correct prices, limits and trial days for both plans
func updatePlanValues(ctx context.Context, db *mongo.Database) error {
	plans := db.Collection("plans")

	updates := []struct {
		name   string
		fields bson.M
	}{
		{
			name: domain.PlanNameInicial,
			fields: bson.M{
				"price":       49.99,
				"max_tables":  10,
				"max_branches": 1,
				"trial_days":  30,
				"updated_at":  time.Now(),
			},
		},
		{
			name: domain.PlanNameProfesional,
			fields: bson.M{
				"price":       99.99,
				"max_tables":  domain.Unlimited,
				"max_branches": domain.Unlimited,
				"trial_days":  30,
				"updated_at":  time.Now(),
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
