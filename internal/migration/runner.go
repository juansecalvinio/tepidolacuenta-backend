package migration

import (
	"context"
	"errors"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Migration represents a single data migration
type Migration struct {
	Name string
	Run  func(ctx context.Context, db *mongo.Database) error
}

// appliedMigration is the record stored in MongoDB
type appliedMigration struct {
	Name      string    `bson:"name"`
	AppliedAt time.Time `bson:"applied_at"`
}

// Runner executes pending migrations and tracks applied ones
type Runner struct {
	db         *mongo.Database
	collection *mongo.Collection
	migrations []Migration
}

// NewRunner creates a new migration runner with the given migrations
func NewRunner(db *mongo.Database, migrations []Migration) *Runner {
	return &Runner{
		db:         db,
		collection: db.Collection("migrations"),
		migrations: migrations,
	}
}

// Run executes all pending migrations in order
func (r *Runner) Run(ctx context.Context) error {
	for _, m := range r.migrations {
		applied, err := r.isApplied(ctx, m.Name)
		if err != nil {
			return err
		}
		if applied {
			continue
		}

		log.Printf("→ Running migration: %s", m.Name)
		if err := m.Run(ctx, r.db); err != nil {
			return err
		}

		if err := r.markApplied(ctx, m.Name); err != nil {
			return err
		}
		log.Printf("✓ Migration applied: %s", m.Name)
	}
	return nil
}

func (r *Runner) isApplied(ctx context.Context, name string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := r.collection.FindOne(ctx, bson.M{"name": name}).Err()
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *Runner) markApplied(ctx context.Context, name string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := r.collection.InsertOne(ctx, appliedMigration{
		Name:      name,
		AppliedAt: time.Now(),
	})
	return err
}
