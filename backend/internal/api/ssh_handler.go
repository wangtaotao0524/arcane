package api

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ofkm/arcane-backend/internal/dto"
	"github.com/ofkm/arcane-backend/internal/middleware"
	"github.com/ofkm/arcane-backend/internal/services"
)

type SSHHandler struct {
	sshService *services.SSHService
}

func NewSSHHandler(
	group *gin.RouterGroup,
	sshService *services.SSHService,
	authMiddleware *middleware.AuthMiddleware,
) {
	h := &SSHHandler{
		sshService: sshService,
	}

	sshGroup := group.Group("/ssh")
	sshGroup.Use(authMiddleware.WithAdminNotRequired().Add())
	{
		sshGroup.POST("/connect", h.Connect)
		sshGroup.GET("/connections", h.ListConnections)
		sshGroup.GET("/connections/:id/status", h.GetConnectionStatus)
		sshGroup.DELETE("/connections/:id", h.Disconnect)
		sshGroup.GET("/terminal/:sessionId", h.TerminalWebSocket)
	}
}

// Connect establishes SSH connection to a remote host
type SSHConnectRequest struct {
	EnvironmentID string `json:"environmentId" binding:"required"`
	Host          string `json:"host" binding:"required"`
	Port          int    `json:"port" binding:"required,min=1,max=65535"`
	Username      string `json:"username" binding:"required"`
	Password      string `json:"password,omitempty"`
	PrivateKey    string `json:"privateKey,omitempty"`
}

func (h *SSHHandler) Connect(c *gin.Context) {
	var req SSHConnectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"data": gin.H{"error": "Invalid request: " + err.Error()},
		})
		return
	}

	var privateKey []byte
	if req.PrivateKey != "" {
		privateKey = []byte(req.PrivateKey)
	}

	conn, err := h.sshService.Connect(c.Request.Context(), req.EnvironmentID, req.Host, req.Port, req.Username, req.Password, privateKey)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "Failed to establish SSH connection",
			slog.String("host", req.Host),
			slog.String("username", req.Username),
			slog.String("error", err.Error()))

		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"data": gin.H{"error": "Failed to connect: " + err.Error()},
		})
		return
	}

	response := dto.SSHConnectionDto{
		ID:           conn.ID,
		EnvironmentID: conn.EnvironmentID,
		Host:         conn.Host,
		Port:         conn.Port,
		Username:     conn.Username,
		Status:       conn.Status,
		CreatedAt:    conn.CreatedAt,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// ListConnections returns all active SSH connections
func (h *SSHHandler) ListConnections(c *gin.Context) {
	connections := h.sshService.ListConnections(c.Request.Context())

	response := make([]dto.SSHConnectionDto, 0, len(connections))
	for _, conn := range connections {
		response = append(response, dto.SSHConnectionDto{
			ID:           conn.ID,
			EnvironmentID: conn.EnvironmentID,
			Host:         conn.Host,
			Port:         conn.Port,
			Username:     conn.Username,
			Status:       conn.Status,
			CreatedAt:    conn.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// GetConnectionStatus returns the status of a specific SSH connection
func (h *SSHHandler) GetConnectionStatus(c *gin.Context) {
	sessionID := c.Param("id")

	conn, err := h.sshService.GetConnectionStatus(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"data": gin.H{"error": "SSH connection not found"},
		})
		return
	}

	response := dto.SSHConnectionDto{
		ID:           conn.ID,
		EnvironmentID: conn.EnvironmentID,
		Host:         conn.Host,
		Port:         conn.Port,
		Username:     conn.Username,
		Status:       conn.Status,
		CreatedAt:    conn.CreatedAt,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// Disconnect terminates an SSH connection
func (h *SSHHandler) Disconnect(c *gin.Context) {
	sessionID := c.Param("id")

	err := h.sshService.Disconnect(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"data": gin.H{"error": "SSH connection not found"},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{"message": "SSH connection terminated successfully"},
	})
}

// TerminalWebSocket handles WebSocket connections for SSH terminal
func (h *SSHHandler) TerminalWebSocket(c *gin.Context) {
	sessionID := c.Param("sessionId")

	err := h.sshService.HandleWebSocket(c.Writer, c.Request, sessionID)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "Failed to handle SSH terminal WebSocket",
			slog.String("sessionId", sessionID),
			slog.String("error", err.Error()))

		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"data": gin.H{"error": "Failed to establish terminal session"},
		})
		return
	}
}