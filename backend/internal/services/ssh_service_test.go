package services

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSSHService_Connect(t *testing.T) {
	service := NewSSHService()
	ctx := context.Background()

	// Test with valid parameters
	conn, err := service.Connect(ctx, "env-123", "localhost", 22, "testuser", "", nil)
	
	// Should return error since we can't connect to local SSH (unless SSH server is running)
	// This test verifies the service handles the connection attempt properly
	assert.Error(t, err)
	assert.Nil(t, conn)
}

func TestSSHService_ListConnections(t *testing.T) {
	service := NewSSHService()
	ctx := context.Background()

	// Initially should have no connections
	connections := service.ListConnections(ctx)
	assert.Empty(t, connections)

	// After adding connections (simulated), should return them
	// Note: We can't actually connect in unit tests, so we test the service logic
}

func TestSSHService_Disconnect(t *testing.T) {
	service := NewSSHService()
	ctx := context.Background()

	// Disconnect non-existent session should return error
	err := service.Disconnect(ctx, "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "SSH connection not found")
}

func TestSSHService_GetConnectionStatus(t *testing.T) {
	service := NewSSHService()
	ctx := context.Background()

	// Get status of non-existent session should return error
	conn, err := service.GetConnectionStatus(ctx, "nonexistent")
	assert.Error(t, err)
	assert.Nil(t, conn)
	assert.Contains(t, err.Error(), "SSH connection not found")
}

func TestGenerateSessionID(t *testing.T) {
	id1 := generateSessionID()
	id2 := generateSessionID()
	
	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2) // Should be different
	assert.Contains(t, id1, "ssh_")
}

func TestGenerateRandomString(t *testing.T) {
	str1 := generateRandomString(8)
	str2 := generateRandomString(8)
	
	assert.Len(t, str1, 8)
	assert.Len(t, str2, 8)
	assert.NotEqual(t, str1, str2) // Should be different
}