package controller

import (
	"net/http"
	"strings"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/bootstrap"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"github.com/gin-gonic/gin"
)

func setAuthCookies(c *gin.Context, env *bootstrap.Env, accessToken string, refreshToken string) {
	if accessToken == "" && refreshToken == "" {
		return
	}

	sameSite := resolveSameSite(env.CookieSameSite)
	c.SetSameSite(sameSite)

	domainName := env.CookieDomain
	secure := env.CookieSecure

	if accessToken != "" {
		c.SetCookie(domain.AccessTokenCookieName, accessToken, hoursToSeconds(env.AccessTokenExpiryHour), "/", domainName, secure, true)
	}

	if refreshToken != "" {
		c.SetCookie(domain.RefreshTokenCookieName, refreshToken, hoursToSeconds(env.RefreshTokenExpiryHour), "/", domainName, secure, true)
	}
}

func resolveSameSite(mode string) http.SameSite {
	switch strings.ToLower(mode) {
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	case "lax":
		fallthrough
	default:
		return http.SameSiteLaxMode
	}
}

func hoursToSeconds(hours int) int {
	if hours <= 0 {
		return 0
	}
	return hours * 3600
}
