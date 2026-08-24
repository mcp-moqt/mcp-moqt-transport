package mcpmoqt

import (
	"crypto/rand"
	"encoding/hex"
	"sync/atomic"
	"time"
)

var sessionCounter atomic.Uint64

func generateSessionID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		ts := time.Now().UnixNano()
		counter := sessionCounter.Add(1)
		return hex.EncodeToString([]byte{
			byte(ts >> 56), byte(ts >> 48), byte(ts >> 40), byte(ts >> 32),
			byte(ts >> 24), byte(ts >> 16), byte(ts >> 8), byte(ts),
			byte(counter >> 56), byte(counter >> 48), byte(counter >> 40), byte(counter >> 32),
			byte(counter >> 24), byte(counter >> 16), byte(counter >> 8), byte(counter),
		})
	}
	return hex.EncodeToString(b)
}
