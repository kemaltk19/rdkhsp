package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

// GenerateRandomToken generates a 32-byte secure random token.
func GenerateRandomToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// HashToken returns the SHA-256 hash of a string token.
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
