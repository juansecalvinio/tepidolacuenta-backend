package middleware

import (
	"strings"

	"juansecalvinio/tepidolacuenta/internal/pkg"

	"github.com/gin-gonic/gin"
)

const (
	authorizationHeader = "Authorization"
	bearerPrefix        = "Bearer "
	userIDKey           = "userId"
	userEmailKey        = "userEmail"
	usernameKey         = "username"
)

// AuthMiddleware creates a middleware for JWT authentication
func AuthMiddleware(jwtService *pkg.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get authorization header
		authHeader := c.GetHeader(authorizationHeader)
		if authHeader == "" {
			pkg.UnauthorizedResponse(c, "Authorization header is required", pkg.ErrUnauthorized)
			c.Abort()
			return
		}

		// Check if it starts with "Bearer "
		if !strings.HasPrefix(authHeader, bearerPrefix) {
			pkg.UnauthorizedResponse(c, "Invalid authorization header format", pkg.ErrUnauthorized)
			c.Abort()
			return
		}

		// Extract token
		tokenString := strings.TrimPrefix(authHeader, bearerPrefix)
		if tokenString == "" {
			pkg.UnauthorizedResponse(c, "Token is required", pkg.ErrUnauthorized)
			c.Abort()
			return
		}

		// Validate token
		claims, err := jwtService.ValidateToken(tokenString)
		if err != nil {
			pkg.UnauthorizedResponse(c, "Invalid or expired token", err)
			c.Abort()
			return
		}

		// Set user information in context
		c.Set(userIDKey, claims.UserID)
		c.Set(userEmailKey, claims.Email)
		c.Set(usernameKey, claims.Username)

		c.Next()
	}
}

// GetUserID gets the user ID from the context
func GetUserID(c *gin.Context) (string, bool) {
	userID, exists := c.Get(userIDKey)
	if !exists {
		return "", false
	}
	return userID.(string), true
}

// GetUserEmail gets the user email from the context
func GetUserEmail(c *gin.Context) (string, bool) {
	email, exists := c.Get(userEmailKey)
	if !exists {
		return "", false
	}
	return email.(string), true
}

// GetUsername gets the username from the context
func GetUsername(c *gin.Context) (string, bool) {
	name, exists := c.Get(usernameKey)
	if !exists {
		return "", false
	}
	return name.(string), true
}
