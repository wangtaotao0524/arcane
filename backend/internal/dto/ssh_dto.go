package dto

import (
	"time"
)

type SSHConnectionDto struct {
	ID           string    `json:"id"`
	EnvironmentID string    `json:"environmentId"`
	Host         string    `json:"host"`
	Port         int       `json:"port"`
	Username     string    `json:"username"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"createdAt"`
}

type SSHConnectRequest struct {
	EnvironmentID string `json:"environmentId" binding:"required"`
	Host          string `json:"host" binding:"required"`
	Port          int    `json:"port" binding:"required,min=1,max=65535"`
	Username      string `json:"username" binding:"required"`
	Password      string `json:"password,omitempty"`
	PrivateKey    string `json:"privateKey,omitempty"`
}

type SSHTerminalSession struct {
	SessionID   string `json:"sessionId"`
	EnvironmentID string `json:"environmentId"`
	Host        string `json:"host"`
	Username    string `json:"username"`
	Active      bool   `json:"active"`
	CreatedAt   time.Time `json:"createdAt"`
}

type SSHTerminalResizeRequest struct {
	Cols int `json:"cols" binding:"required"`
	Rows int `json:"rows" binding:"required"`
}