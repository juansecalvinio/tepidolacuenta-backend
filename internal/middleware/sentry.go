package middleware

import (
	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
)

// SentryContext enriquece el scope de Sentry con el usuario y el local del JWT,
// para poder filtrar/alertar por usuario, rol y sucursal en los eventos.
// Debe registrarse DESPUÉS de AuthMiddleware (necesita los claims en el contexto).
func SentryContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		hub := sentrygin.GetHubFromContext(c)
		if hub == nil {
			c.Next()
			return
		}

		userID, _ := GetUserID(c)
		email, _ := GetUserEmail(c)
		hub.Scope().SetUser(sentry.User{ID: userID, Email: email})

		if role, ok := GetUserRole(c); ok {
			hub.Scope().SetTag("role", role)
		}
		if restaurantID, ok := GetUserRestaurantID(c); ok {
			hub.Scope().SetTag("restaurantId", restaurantID)
		}
		if branchID, ok := GetUserBranchID(c); ok {
			hub.Scope().SetTag("branchId", branchID)
		}

		c.Next()
	}
}
