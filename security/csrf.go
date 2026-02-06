package security

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// CSRFToken represents a CSRF token for a session
type CSRFToken struct {
	Token     string
	ExpiresAt time.Time
}

// CSRFManager manages CSRF tokens
type CSRFManager struct {
	tokens map[string]*CSRFToken
	mu     sync.RWMutex
}

// Global CSRF manager
var Manager = &CSRFManager{
	tokens: make(map[string]*CSRFToken),
}

// GenerateToken creates a new CSRF token for a session
func (m *CSRFManager) GenerateToken(sessionID string) (string, error) {
	// Generate 32 random bytes
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate CSRF token: %w", err)
	}

	token := base64.URLEncoding.EncodeToString(bytes)

	m.mu.Lock()
	defer m.mu.Unlock()

	m.tokens[sessionID] = &CSRFToken{
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	return token, nil
}

// ValidateToken checks if a CSRF token is valid
func (m *CSRFManager) ValidateToken(sessionID, token string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	csrfToken, exists := m.tokens[sessionID]
	if !exists {
		return false
	}

	// Check if expired
	if time.Now().After(csrfToken.ExpiresAt) {
		// Clean up expired token
		go m.DeleteToken(sessionID)
		return false
	}

	return csrfToken.Token == token
}

// GetToken retrieves the CSRF token for a session
func (m *CSRFManager) GetToken(sessionID string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if token, exists := m.tokens[sessionID]; exists {
		if time.Now().Before(token.ExpiresAt) {
			return token.Token
		}
	}
	return ""
}

// DeleteToken removes a CSRF token
func (m *CSRFManager) DeleteToken(sessionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.tokens, sessionID)
}

// Middleware adds CSRF token to request context
func Middleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get session cookie
		cookie, err := r.Cookie("session_id")
		if err != nil {
			next(w, r)
			return
		}

		sessionID := cookie.Value

		// For GET requests, ensure token exists
		if r.Method == "GET" {
			if Manager.GetToken(sessionID) == "" {
				_, _ = Manager.GenerateToken(sessionID)
			}
		}

		// For POST requests, validate token
		if r.Method == "POST" {
			csrfToken := r.FormValue("csrf_token")
			if csrfToken == "" {
				http.Error(w, "CSRF token missing", http.StatusForbidden)
				return
			}

			if !Manager.ValidateToken(sessionID, csrfToken) {
				http.Error(w, "Invalid CSRF token", http.StatusForbidden)
				return
			}
		}

		next(w, r)
	}
}
