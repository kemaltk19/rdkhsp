package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
	"os"
)

// encKey derives a 32-byte AES-256 key from APP_ENCRYPTION_KEY (or JWT_SECRET).
func encKey() []byte {
	k := os.Getenv("APP_ENCRYPTION_KEY")
	if k == "" {
		k = os.Getenv("JWT_SECRET")
	}
	if k == "" {
		k = "radikal-hesap-default-dev-key-change-me"
	}
	sum := sha256.Sum256([]byte(k))
	return sum[:]
}

// Encrypt returns base64(nonce||ciphertext) using AES-GCM. Empty in -> empty out.
func Encrypt(plain string) (string, error) {
	if plain == "" {
		return "", nil
	}
	block, err := aes.NewCipher(encKey())
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ct := gcm.Seal(nonce, nonce, []byte(plain), nil)
	return base64.StdEncoding.EncodeToString(ct), nil
}

// Decrypt reverses Encrypt.
func Decrypt(enc string) (string, error) {
	if enc == "" {
		return "", nil
	}
	data, err := base64.StdEncoding.DecodeString(enc)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(encKey())
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	ns := gcm.NonceSize()
	if len(data) < ns {
		return "", errors.New("ciphertext too short")
	}
	nonce, ct := data[:ns], data[ns:]
	pt, err := gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		return "", err
	}
	return string(pt), nil
}
