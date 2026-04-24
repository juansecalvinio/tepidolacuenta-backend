package pkg

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	ErrInvalidTokenClaims = errors.New("invalid token claims")
	ErrTokenExpired       = errors.New("token has expired")
)

// Claims represents the JWT claims
type Claims struct {
	UserID       string `json:"userId"`
	Email        string `json:"email"`
	Role         string `json:"role"`
	RestaurantID string `json:"restaurantId,omitempty"`
	jwt.RegisteredClaims
}

// JWTService handles JWT token operations
type JWTService struct {
	secretKey []byte
}

// NewJWTService creates a new JWT service
func NewJWTService(secretKey string) *JWTService {
	return &JWTService{
		secretKey: []byte(secretKey),
	}
}

// GenerateToken generates a new JWT token for a user.
// role should be "owner" or "employee". restaurantID should be "" for owners.
func (s *JWTService) GenerateToken(userID primitive.ObjectID, email, role, restaurantID string) (string, error) {
	claims := &Claims{
		UserID:       userID.Hex(),
		Email:        email,
		Role:         role,
		RestaurantID: restaurantID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secretKey)
}

// ValidateToken validates a JWT token and returns the claims
func (s *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return s.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidTokenClaims
	}

	return claims, nil
}

// GetUserIDFromToken extracts the user ID from a token string
func (s *JWTService) GetUserIDFromToken(tokenString string) (primitive.ObjectID, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return primitive.NilObjectID, err
	}

	userID, err := primitive.ObjectIDFromHex(claims.UserID)
	if err != nil {
		return primitive.NilObjectID, ErrInvalidTokenClaims
	}

	return userID, nil
}
