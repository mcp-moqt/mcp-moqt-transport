package mcpmoqt

import (
	"crypto/rand"
	"encoding/hex"
)

// generateSessionID generates a random session ID.
func generateSessionID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		// Fallback to a deterministic value if crypto/rand fails
		return "fallback-session-id"
	}
	return hex.EncodeToString(b)
}
