package repository

import (
	"context"
	"time"

	"juansecalvinio/tepidolacuenta/internal/invitation/domain"
	"juansecalvinio/tepidolacuenta/internal/pkg"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type mongoRepository struct {
	collection *mongo.Collection
}

func NewMongoRepository(db *mongo.Database) Repository {
	return &mongoRepository{
		collection: db.Collection("invitations"),
	}
}

func (r *mongoRepository) Create(ctx context.Context, inv *domain.Invitation) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := r.collection.InsertOne(ctx, inv)
	if err != nil {
		return err
	}

	inv.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *mongoRepository) FindByCode(ctx context.Context, code string) (*domain.Invitation, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var inv domain.Invitation
	err := r.collection.FindOne(ctx, bson.M{"code": code}).Decode(&inv)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, pkg.ErrNotFound
		}
		return nil, err
	}

	return &inv, nil
}

func (r *mongoRepository) MarkUsed(ctx context.Context, code string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"code": code},
		bson.M{"$set": bson.M{"used": true}},
	)
	return err
}
