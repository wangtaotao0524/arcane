package api

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ofkm/arcane-backend/internal/dto"
	"github.com/ofkm/arcane-backend/internal/middleware"
	"github.com/ofkm/arcane-backend/internal/services"
)

type AuthHandler struct {
	userService *services.UserService
	authService *services.AuthService
	oidcService *services.OidcService
}

func NewAuthHandler(userService *services.UserService, authService *services.AuthService, oidcService *services.OidcService) *AuthHandler {
	return &AuthHandler{
		userService: userService,
		authService: authService,
		oidcService: oidcService,
	}
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Success               bool                 `json:"success"`
	Token                 string               `json:"token,omitempty"`
	RefreshToken          string               `json:"refreshToken,omitempty"`
	ExpiresAt             time.Time            `json:"expiresAt,omitempty"`
	User                  *dto.UserResponseDto `json:"user,omitempty"`
	Error                 string               `json:"error,omitempty"`
	RequirePasswordChange bool                 `json:"requirePasswordChange,omitempty"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

type OidcLoginRequest struct {
	Code  string `json:"code" binding:"required"`
	State string `json:"state" binding:"required"`
}

type PasswordChangeRequest struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"data":    gin.H{"error": "Invalid request format"},
		})
		return
	}

	localAuthEnabled, err := h.authService.IsLocalAuthEnabled(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"data":    gin.H{"error": "Failed to check authentication settings"},
		})
		return
	}

	if !localAuthEnabled {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"data":    gin.H{"error": "Local authentication is disabled"},
		})
		return
	}

	user, tokenPair, err := h.authService.Login(c.Request.Context(), req.Username, req.Password)

	// Handle password change required
	if errors.Is(err, services.ErrPasswordChangeRequired) && user != nil {
		if tokenPair != nil {
			c.SetSameSite(http.SameSiteLaxMode)
			maxAge := int(time.Until(tokenPair.ExpiresAt).Seconds())
			if maxAge < 0 {
				maxAge = 0
			}
			secure := c.Request.TLS != nil
			c.SetCookie("token", tokenPair.AccessToken, maxAge, "/", "", secure, true)
		}

		var out dto.UserResponseDto
		if mapErr := dto.MapStruct(user, &out); mapErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"data":    gin.H{"error": "Failed to map user"},
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"requirePasswordChange": true,
				"token": func() string {
					if tokenPair != nil {
						return tokenPair.AccessToken
					}
					return ""
				}(),
				"refreshToken": func() string {
					if tokenPair != nil {
						return tokenPair.RefreshToken
					}
					return ""
				}(),
				"expiresAt": func() time.Time {
					if tokenPair != nil {
						return tokenPair.ExpiresAt
					}
					return time.Time{}
				}(),
				"user": out,
			},
		})
		return
	}

	if err != nil {
		var statusCode int
		var errorMsg string

		switch {
		case errors.Is(err, services.ErrInvalidCredentials):
			statusCode = http.StatusUnauthorized
			errorMsg = "Invalid username or password"
		case errors.Is(err, services.ErrLocalAuthDisabled):
			statusCode = http.StatusBadRequest
			errorMsg = "Local authentication is disabled"
		default:
			statusCode = http.StatusInternalServerError
			errorMsg = "Authentication failed"
		}

		c.JSON(statusCode, gin.H{
			"success": false,
			"data":    gin.H{"error": errorMsg},
		})
		return
	}

	// Set token cookie
	c.SetSameSite(http.SameSiteLaxMode)
	maxAge := int(time.Until(tokenPair.ExpiresAt).Seconds())
	if maxAge < 0 {
		maxAge = 0
	}
	secure := c.Request.TLS != nil
	c.SetCookie("token", tokenPair.AccessToken, maxAge, "/", "", secure, true)

	var out dto.UserResponseDto
	if mapErr := dto.MapStruct(user, &out); mapErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"data":    gin.H{"error": "Failed to map user"},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"token":        tokenPair.AccessToken,
			"refreshToken": tokenPair.RefreshToken,
			"expiresAt":    tokenPair.ExpiresAt,
			"user":         out,
		},
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("token", "", -1, "/", "", false, true)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    gin.H{"message": "Logged out successfully"},
	})
}

func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"data":    gin.H{"error": "Not authenticated"},
		})
		return
	}

	user, err := h.userService.GetUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"data":    gin.H{"error": "Failed to get user information"},
		})
		return
	}

	var out dto.UserResponseDto
	if mapErr := dto.MapStruct(user, &out); mapErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"data":    gin.H{"error": "Failed to map user"},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    out,
	})
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"data":    gin.H{"error": "Invalid request format"},
		})
		return
	}

	tokenPair, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		var statusCode int
		var errorMsg string

		switch {
		case errors.Is(err, services.ErrInvalidToken), errors.Is(err, services.ErrExpiredToken):
			statusCode = http.StatusUnauthorized
			errorMsg = "Invalid or expired refresh token"
		default:
			statusCode = http.StatusInternalServerError
			errorMsg = "Failed to refresh token"
		}

		c.JSON(statusCode, gin.H{
			"success": false,
			"data":    gin.H{"error": errorMsg},
		})
		return
	}

	c.SetSameSite(http.SameSiteLaxMode)
	maxAge := int(time.Until(tokenPair.ExpiresAt).Seconds())
	if maxAge < 0 {
		maxAge = 0
	}
	secure := c.Request.TLS != nil
	c.SetCookie("token", tokenPair.AccessToken, maxAge, "/", "", secure, true)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"token":        tokenPair.AccessToken,
			"refreshToken": tokenPair.RefreshToken,
			"expiresAt":    tokenPair.ExpiresAt,
		},
	})
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"data":    gin.H{"error": "Not authenticated"},
		})
		return
	}

	var req PasswordChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"data":    gin.H{"error": "Invalid request format"},
		})
		return
	}

	if !user.RequirePasswordChange && req.CurrentPassword == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"data":    gin.H{"error": "Current password is required"},
		})
		return
	}

	err := h.authService.ChangePassword(
		c.Request.Context(),
		user.ID,
		req.CurrentPassword,
		req.NewPassword,
	)

	if err != nil {
		var statusCode int
		var errorMsg string

		switch {
		case errors.Is(err, services.ErrInvalidCredentials):
			statusCode = http.StatusUnauthorized
			errorMsg = "Current password is incorrect"
		default:
			statusCode = http.StatusInternalServerError
			errorMsg = "Failed to change password"
		}

		c.JSON(statusCode, gin.H{
			"success": false,
			"data":    gin.H{"error": errorMsg},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    gin.H{"message": "Password changed successfully"},
	})
}
