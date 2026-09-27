package controller

import (
	"crypto/rand"
	"encoding/hex"
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
	if sameSite == http.SameSiteNoneMode && !secure {
		// SameSite=None requires the Secure attribute to avoid browser rejection.
		secure = true
	}

	if accessToken != "" {
		c.SetCookie(domain.AccessTokenCookieName, accessToken, hoursToSeconds(env.AccessTokenExpiryHour), "/", domainName, secure, true)
	}

	if refreshToken != "" {
		c.SetCookie(domain.RefreshTokenCookieName, refreshToken, hoursToSeconds(env.RefreshTokenExpiryHour), "/", domainName, secure, true)
	}
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err == nil {
		c.SetCookie(domain.CSRFTokenCookieName, hex.EncodeToString(tokenBytes), hoursToSeconds(env.RefreshTokenExpiryHour), "/", domainName, secure, false)
	}
}

func setSessionRoleCookie(c *gin.Context, env *bootstrap.Env, role domain.UserType) {
	sameSite := resolveSameSite(env.CookieSameSite)
	c.SetSameSite(sameSite)
	secure := env.CookieSecure || sameSite == http.SameSiteNoneMode
	c.SetCookie(domain.SessionRoleCookieName, string(role), hoursToSeconds(env.AccessTokenExpiryHour), "/", env.CookieDomain, secure, false)
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
