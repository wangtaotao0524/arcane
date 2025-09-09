package cookie

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

var (
	TokenCookieName     = "__Host-token" // #nosec G101: cookie name label, not a credential
	OidcStateCookieName = "oidc_state"
)

func CreateTokenCookie(c *gin.Context, maxAgeInSeconds int, token string) {
	if maxAgeInSeconds < 0 {
		maxAgeInSeconds = 0
	}
	c.SetCookie(TokenCookieName, token, maxAgeInSeconds, "/", "", true, true)
}

func ClearTokenCookie(c *gin.Context) {
	c.SetCookie(TokenCookieName, "", -1, "/", "", true, true)
}

func CreateOidcStateCookie(c *gin.Context, value string, maxAgeInSeconds int) {
	c.SetSameSite(http.SameSiteLaxMode)
	if maxAgeInSeconds < 0 {
		maxAgeInSeconds = 0
	}
	c.SetCookie(OidcStateCookieName, value, maxAgeInSeconds, "/", "", true, true)
}

func GetOidcStateCookie(c *gin.Context) (string, error) {
	return c.Cookie(OidcStateCookieName)
}

func ClearOidcStateCookie(c *gin.Context) {
	c.SetCookie(OidcStateCookieName, "", -1, "/", "", true, true)
}
