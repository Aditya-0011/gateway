package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"unsafe"
)

func GenerateSessionKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

func HashSHA256(data string) string {
	hash := sha256.Sum256(unsafe.Slice(unsafe.StringData(data), len(data)))
	return hex.EncodeToString(hash[:])
}
