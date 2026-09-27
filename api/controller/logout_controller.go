package controller

import (
	"net/http"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/bootstrap"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"github.com/gin-gonic/gin"
)

type LogoutController struct {
	Env *bootstrap.Env
}

func (lc *LogoutController) Logout(c *gin.Context) {
	sameSite := resolveSameSite(lc.Env.CookieSameSite)
	c.SetSameSite(sameSite)
	secure := lc.Env.CookieSecure || sameSite == http.SameSiteNoneMode
	domainName := lc.Env.CookieDomain
	c.SetCookie(domain.AccessTokenCookieName, "", -1, "/", domainName, secure, true)
	c.SetCookie(domain.RefreshTokenCookieName, "", -1, "/", domainName, secure, true)
	c.SetCookie(domain.SessionRoleCookieName, "", -1, "/", domainName, secure, false)
	c.SetCookie(domain.CSRFTokenCookieName, "", -1, "/", domainName, secure, false)
	c.JSON(http.StatusOK, domain.SuccessResponse{Message: "logged out"})
}
