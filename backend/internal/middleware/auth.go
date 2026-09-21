package middleware

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/nwasiq/fieldops/backend/internal/services"
)

// Authenticate requires a valid bearer token and puts the user id and role on
// the context.
func Authenticate(auth *services.AuthService, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			abort(c, http.StatusUnauthorized, "missing bearer token")
			return
		}
		user, err := auth.Authenticate(c.Request.Context(), strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			if services.KindOf(err) == services.KindUnauthorized {
				abort(c, http.StatusUnauthorized, err.Error())
				return
			}
			AbortInternal(c, logger, err)
			return
		}
		SetAuthenticatedUser(c, user.ID, user.Role)
		c.Next()
	}
}

// RequireRole refuses any authenticated user whose role is not listed (§1.1).
func RequireRole(roles ...string) gin.HandlerFunc {
	allowed := map[string]struct{}{}
	for _, role := range roles {
		allowed[role] = struct{}{}
	}
	return func(c *gin.Context) {
		if _, ok := allowed[Role(c)]; !ok {
			abort(c, http.StatusForbidden, "insufficient role")
			return
		}
		c.Next()
	}
}
