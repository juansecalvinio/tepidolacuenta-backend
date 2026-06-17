package middleware

import (
	"juansecalvinio/tepidolacuenta/internal/auth/domain"
	"juansecalvinio/tepidolacuenta/internal/pkg"

	"github.com/gin-gonic/gin"
)

// OwnerOnly rejects requests whose JWT role is not "owner".
// Must be used after AuthMiddleware.
func OwnerOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := GetUserRole(c)
		if !exists || role != string(domain.RoleOwner) {
			pkg.ForbiddenResponse(c, "This action requires owner privileges", pkg.ErrForbidden)
			c.Abort()
			return
		}
		c.Next()
	}
}
