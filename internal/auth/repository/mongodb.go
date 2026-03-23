package repository

import (
	"context"
	"errors"
	"time"

	"juansecalvinio/tepidolacuenta/internal/auth/domain"
	"juansecalvinio/tepidolacuenta/internal/pkg"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type mongoRepository struct {
	collection *mongo.Collection
}

// NewMongoRepository creates a new MongoDB repository for users
func NewMongoRepository(db *mongo.Database) Repository {
	return &mongoRepository{
		collection: db.Collection("users"),
	}
}

func (r *mongoRepository) Create(ctx context.Context, user *domain.User) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Check if user already exists
	existing, _ := r.FindByEmail(ctx, user.Email)
	if existing != nil {
		return pkg.ErrUserAlreadyExists
	}

	result, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		return err
	}

	user.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *mongoRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var user domain.User
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, pkg.ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (r *mongoRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var user domain.User
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, pkg.ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (r *mongoRepository) FindByGoogleID(ctx context.Context, googleID string) (*domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var user domain.User
	err := r.collection.FindOne(ctx, bson.M{"google_id": googleID}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, pkg.ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (r *mongoRepository) LinkGoogleID(ctx context.Context, id primitive.ObjectID, googleID string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"google_id":  googleID,
			"updated_at": time.Now(),
		},
	}

	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	return err
}

func (r *mongoRepository) Update(ctx context.Context, user *domain.User) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	user.UpdatedAt = time.Now()

	update := bson.M{
		"$set": bson.M{
			"email":      user.Email,
			"updated_at": user.UpdatedAt,
		},
	}

	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": user.ID}, update)
	return err
}

func (r *mongoRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *mongoRepository) SaveResetToken(ctx context.Context, id primitive.ObjectID, token string, expiry time.Time) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"reset_password_token":  token,
			"reset_password_expiry": expiry,
			"updated_at":            time.Now(),
		},
	}

	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	return err
}

func (r *mongoRepository) FindByResetToken(ctx context.Context, token string) (*domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var user domain.User
	err := r.collection.FindOne(ctx, bson.M{"reset_password_token": token}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, pkg.ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (r *mongoRepository) UpdatePasswordAndClearToken(ctx context.Context, id primitive.ObjectID, hashedPassword string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"password":   hashedPassword,
			"updated_at": time.Now(),
		},
		"$unset": bson.M{
			"reset_password_token":  "",
			"reset_password_expiry": "",
		},
	}

	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	return err
}
