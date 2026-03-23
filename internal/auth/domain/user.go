package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID                   primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Email                string             `json:"email" bson:"email"`
	Password             string             `json:"-" bson:"password,omitempty"`
	GoogleID             string             `json:"-" bson:"google_id,omitempty"`
	ResetPasswordToken   string             `json:"-" bson:"reset_password_token,omitempty"`
	ResetPasswordExpiry  time.Time          `json:"-" bson:"reset_password_expiry,omitempty"`
	CreatedAt            time.Time          `json:"createdAt" bson:"created_at"`
	UpdatedAt            time.Time          `json:"updatedAt" bson:"updated_at"`
}

// GoogleUserInfo holds the user profile returned by Google's userinfo API
type GoogleUserInfo struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

// RegisterInput represents the data needed to register a new user
type RegisterInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
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

// ForgotPasswordInput represents the data needed to request a password reset
type ForgotPasswordInput struct {
	Email string `json:"email" binding:"required,email"`
}

// ResetPasswordInput represents the data needed to reset the password
type ResetPasswordInput struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
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
func NewUser(email, password string) *User {
	now := time.Now()
	return &User{
		Email:     email,
		Password:  password,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// NewGoogleUser creates a new user authenticated via Google OAuth
func NewGoogleUser(email, googleID string) *User {
	now := time.Now()
	return &User{
		Email:     email,
		GoogleID:  googleID,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
