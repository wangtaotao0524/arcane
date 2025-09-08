package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ofkm/arcane-backend/internal/models"
	"github.com/ofkm/arcane-backend/internal/services"
)

type AuthOptions struct {
	AdminRequired   bool
	SuccessOptional bool
}

type AuthMiddleware struct {
	authService *services.AuthService
	options     AuthOptions
}

func NewAuthMiddleware(authService *services.AuthService) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
		options: AuthOptions{
			AdminRequired:   false,
			SuccessOptional: false,
		},
	}
}

func (m *AuthMiddleware) WithAdminRequired() *AuthMiddleware {
	clone := *m
	clone.options.AdminRequired = true
	return &clone
}

func (m *AuthMiddleware) WithAdminNotRequired() *AuthMiddleware {
	clone := *m
	clone.options.AdminRequired = false
	return &clone
}

func (m *AuthMiddleware) WithSuccessOptional() *AuthMiddleware {
	clone := *m
	clone.options.SuccessOptional = true
	return &clone
}

func (m *AuthMiddleware) Add() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractBearerOrCookieToken(c)
		if token == "" {
			if m.options.SuccessOptional {
				c.Next()
				return
			}
			c.JSON(http.StatusUnauthorized, models.APIError{
				Code:    models.APIErrorCodeUnauthorized,
				Message: "Authentication required",
			})
			c.Abort()
			return
		}

		user, err := m.authService.VerifyToken(c.Request.Context(), token)
		if err != nil {
			if m.options.SuccessOptional {
				c.Next()
				return
			}
			c.JSON(http.StatusUnauthorized, models.APIError{
				Code:    models.APIErrorCodeUnauthorized,
				Message: "Invalid or expired token",
			})
			c.Abort()
			return
		}

		isAdmin := userHasRole(user, "admin")
		if m.options.AdminRequired && !isAdmin {
			if m.options.SuccessOptional {
				c.Next()
				return
			}
			c.JSON(http.StatusForbidden, models.APIError{
				Code:    "FORBIDDEN",
				Message: "You don't have permission to access this resource",
			})
			c.Abort()
			return
		}

		c.Set("userID", user.ID)
		c.Set("currentUser", user)
		c.Set("userIsAdmin", isAdmin)
		c.Next()
	}
}

func extractBearerOrCookieToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}
	if tokenCookie, err := c.Cookie("token"); err == nil && tokenCookie != "" {
		return tokenCookie
	}
	return ""
}

func userHasRole(user *models.User, role string) bool {
	for _, r := range user.Roles {
		if r == role {
			return true
		}
	}
	return false
}

func GetCurrentUserID(c *gin.Context) (string, bool) {
	userID, exists := c.Get("userID")
	if !exists {
		return "", false
	}
	userIDStr, ok := userID.(string)
	return userIDStr, ok
}

func GetCurrentUser(c *gin.Context) (*models.User, bool) {
	user, exists := c.Get("currentUser")
	if !exists {
		return nil, false
	}
	u, ok := user.(*models.User)
	return u, ok
}

func IsAuthenticated(c *gin.Context) bool {
	_, exists := GetCurrentUser(c)
	return exists
}

func RequireAuthentication(c *gin.Context) (*models.User, bool) {
	user, exists := GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIError{
			Code:    models.APIErrorCodeUnauthorized,
			Message: "Authentication required",
		})
		c.Abort()
		return nil, false
	}
	return user, true
}
