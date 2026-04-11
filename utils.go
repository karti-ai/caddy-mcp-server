package main

import (
	"crypto/rand"
	"encoding/base64"
)

func generateSalt() (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	return base64.RawStdEncoding.EncodeToString(salt), nil
}
