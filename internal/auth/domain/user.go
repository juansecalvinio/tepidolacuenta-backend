package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID             primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Email          string             `json:"email" bson:"email"`
	Password       string             `json:"-" bson:"password"`
	RestaurantName string             `json:"restaurantName" bson:"restaurant_name"`
	CreatedAt      time.Time          `json:"createdAt" bson:"created_at"`
	UpdatedAt      time.Time          `json:"updatedAt" bson:"updated_at"`
}

// RegisterInput represents the data needed to register a new user
type RegisterInput struct {
	Email          string `json:"email" binding:"required,email"`
	Password       string `json:"password" binding:"required,min=8"`
	RestaurantName string `json:"restaurantName" binding:"required,min=3"`
}

// LoginInput represents the data needed to login
type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents the response after successful login
type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// HashPassword hashes the user's password using bcrypt
func (u *User) HashPassword() error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

// ComparePassword compares the provided password with the user's hashed password
func (u *User) ComparePassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
}

// NewUser creates a new user with the current timestamp
func NewUser(email, password, restaurantName string) *User {
	now := time.Now()
	return &User{
		Email:          email,
		Password:       password,
		RestaurantName: restaurantName,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}
