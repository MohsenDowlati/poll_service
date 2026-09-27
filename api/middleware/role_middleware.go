package middleware

import (
	"net/http"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"github.com/gin-gonic/gin"
)

// RequireRoles restricts a route to one of the supplied roles. It must run
// after JwtAuthMiddleware so the role has already been validated and stored.
func RequireRoles(roles ...domain.UserType) gin.HandlerFunc {
	allowed := make(map[domain.UserType]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}

	return func(c *gin.Context) {
		role := domain.UserType(c.GetString("x-user-type"))
		if _, ok := allowed[role]; !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, domain.ErrorResponse{Message: "forbidden"})
			return
		}
		c.Next()
	}
}
