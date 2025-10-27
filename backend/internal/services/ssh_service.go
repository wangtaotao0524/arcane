package services

import (
	"context"
	"crypto/rand"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"
)

type SSHConnection struct {
	ID           string
	EnvironmentID string
	Host         string
	Port         int
	Username     string
	SessionID    string
	CreatedAt    time.Time
	Status       string // "connected", "disconnected", "error"
}

type SSHService struct {
	connections map[string]*SSHConnection
	mu          sync.RWMutex
	upgrader    websocket.Upgrader
}

func NewSSHService() *SSHService {
	return &SSHService{
		connections: make(map[string]*SSHConnection),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // In production, implement proper origin checking
			},
		},
	}
}

// Connect establishes SSH connection to a remote host
func (s *SSHService) Connect(ctx context.Context, envID, host string, port int, username, password string, privateKey []byte) (*SSHConnection, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	conn := &SSHConnection{
		ID:           generateSessionID(),
		EnvironmentID: envID,
		Host:         host,
		Port:         port,
		Username:     username,
		CreatedAt:    time.Now(),
		Status:       "connecting",
	}

	// Configure SSH client
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // In production, use proper host key verification
		Timeout:        30 * time.Second,
	}

	// Add authentication method
	if len(privateKey) > 0 {
		signer, err := ssh.ParsePrivateKey(privateKey)
		if err != nil {
			return nil, err
		}
		config.Auth = append(config.Auth, ssh.PublicKeys(signer))
	} else if password != "" {
		config.Auth = append(config.Auth, ssh.Password(password))
	}

	// Establish SSH connection
	sshConn, err := ssh.Dial("tcp", net.JoinHostPort(host, strconv.Itoa(port)), config)
	if err != nil {
		conn.Status = "error"
		return nil, err
	}

	conn.Status = "connected"
	s.connections[conn.ID] = conn

	// Start session monitoring
	go s.monitorConnection(ctx, conn, sshConn)

	return conn, nil
}

// HandleWebSocket handles WebSocket connections for SSH terminal
func (s *SSHService) HandleWebSocket(w http.ResponseWriter, r *http.Request, sessionID string) error {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}
	defer conn.Close()

	s.mu.RLock()
	sshConn, exists := s.connections[sessionID]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("SSH session not found")
	}

	// Create SSH session
	sshSession, err := s.createSSHSession(sshConn)
	if err != nil {
		return err
	}
	defer sshSession.Close()

	// Handle terminal session
	return s.handleTerminalSession(conn, sshSession)
}

// ListConnections returns all active SSH connections
func (s *SSHService) ListConnections(ctx context.Context) []*SSHConnection {
	s.mu.RLock()
	defer s.mu.RUnlock()

	connections := make([]*SSHConnection, 0, len(s.connections))
	for _, conn := range s.connections {
		connections = append(connections, conn)
	}

	return connections
}

// Disconnect terminates an SSH connection
func (s *SSHService) Disconnect(ctx context.Context, sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	conn, exists := s.connections[sessionID]
	if !exists {
		return fmt.Errorf("SSH connection not found")
	}

	conn.Status = "disconnected"
	delete(s.connections, sessionID)

	return nil
}

// GetConnectionStatus returns the status of a specific connection
func (s *SSHService) GetConnectionStatus(ctx context.Context, sessionID string) (*SSHConnection, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	conn, exists := s.connections[sessionID]
	if !exists {
		return nil, fmt.Errorf("SSH connection not found")
	}

	return conn, nil
}

// monitorConnection monitors SSH connection health
func (s *SSHService) monitorConnection(ctx context.Context, conn *SSHConnection, sshConn *ssh.Client) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Check if connection is still alive
			_, _, err := sshConn.SendRequest("keepalive@openssh.com", true, nil)
			if err != nil {
				conn.Status = "error"
				return
			}
		}
	}
}

// createSSHSession creates a new SSH session for terminal interaction
func (s *SSHService) createSSHSession(conn *SSHConnection) (*ssh.Session, error) {
	// Configure SSH client
	config := &ssh.ClientConfig{
		User: conn.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(""), // Empty password for now - will be implemented with actual auth
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // In production, use proper host key verification
		Timeout:         30 * time.Second,
	}

	// Establish SSH connection
	sshConn, err := ssh.Dial("tcp", net.JoinHostPort(conn.Host, strconv.Itoa(conn.Port)), config)
	if err != nil {
		return nil, err
	}

	// Create SSH session
	session, err := sshConn.NewSession()
	if err != nil {
		sshConn.Close()
		return nil, err
	}

	return session, nil
}

// handleTerminalSession handles WebSocket to SSH terminal communication
func (s *SSHService) handleTerminalSession(wsConn *websocket.Conn, sshSession *ssh.Session) error {
	// Set up terminal modes
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,     // enable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	// Request pseudo terminal
	if err := sshSession.RequestPty("xterm", 40, 80, modes); err != nil {
		return err
	}

	// Set up pipes
	stdin, err := sshSession.StdinPipe()
	if err != nil {
		return err
	}

	stdout, err := sshSession.StdoutPipe()
	if err != nil {
		return err
	}

	stderr, err := sshSession.StderrPipe()
	if err != nil {
		return err
	}

	// Start shell
	if err := sshSession.Shell(); err != nil {
		return err
	}

	// Handle WebSocket messages
	go func() {
		for {
			_, message, err := wsConn.ReadMessage()
			if err != nil {
				break
			}
			stdin.Write(message)
		}
	}()

	// Handle stdout
	go func() {
		buffer := make([]byte, 4096)
		for {
			n, err := stdout.Read(buffer)
			if err != nil {
				break
			}
			if n > 0 {
				wsConn.WriteMessage(websocket.TextMessage, buffer[:n])
			}
		}
	}()

	// Handle stderr
	go func() {
		buffer := make([]byte, 4096)
		for {
			n, err := stderr.Read(buffer)
			if err != nil {
				break
			}
			if n > 0 {
				wsConn.WriteMessage(websocket.TextMessage, buffer[:n])
			}
		}
	}()

	// Wait for session to end
	err = sshSession.Wait()
	return err
}

func generateSessionID() string {
	return fmt.Sprintf("ssh_%d_%s", time.Now().Unix(), generateRandomString(8))
}

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		// Fallback to timestamp-based ID if crypto fails
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}
	return string(b)
}