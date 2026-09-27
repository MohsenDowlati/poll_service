package middleware

import (
	"net/http"
	"strings"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/internal/tokenutil"
	"github.com/gin-gonic/gin"
)

func JwtAuthMiddleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authToken := extractTokenFromRequest(c)
		if authToken == "" {
			c.JSON(http.StatusUnauthorized, domain.ErrorResponse{Message: "Not authorized"})
			c.Abort()
			return
		}

		authorized, err := tokenutil.IsAuthorized(authToken, secret)
		if !authorized || err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, domain.ErrorResponse{Message: "not authorized"})
			return
		}

		userID, err := tokenutil.ExtractIDFromToken(authToken, secret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, domain.ErrorResponse{Message: "not authorized"})
			return
		}
		userType, err := tokenutil.ExtractRoleFromToken(authToken, secret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, domain.ErrorResponse{Message: "not authorized"})
			return
		}
		c.Set("x-user-id", userID)
		c.Set("x-user-type", userType)
		c.Next()
	}
}

func extractTokenFromRequest(c *gin.Context) string {
	if cookieToken, err := c.Cookie(domain.AccessTokenCookieName); err == nil && cookieToken != "" {
		return cookieToken
	}

	authHeader := c.Request.Header.Get("Authorization")
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) == 2 && strings.EqualFold(parts[0], "bearer") {
		return strings.TrimSpace(parts[1])
	}

	return ""
}
