package utils

import (
	"crypto/rand"
	"encoding/hex"
)

func GenerateUniqueID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		panic("failed to generate user ID")
	}
	return hex.EncodeToString(bytes)
}

func GenerateSecret() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		panic("failed to generate secret")
	}
	return hex.EncodeToString(bytes)
}
