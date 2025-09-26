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
		if authorized {
			userID, err := tokenutil.ExtractIDFromToken(authToken, secret)
			if err != nil {
				c.JSON(http.StatusUnauthorized, domain.ErrorResponse{Message: err.Error()})
				c.Abort()
				return
			}
			c.Set("x-user-id", userID)
			userType, err := tokenutil.ExtractRoleFromToken(authToken, secret)
			if err != nil {
				c.JSON(http.StatusUnauthorized, domain.ErrorResponse{Message: err.Error()})
				c.Abort()
				return
			}
			c.Set("x-user-type", userType)
			c.Next()
			return
		}
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse{Message: err.Error()})
		c.Abort()
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
