// Package utils provides utility functions for the application.
package utils

import (
	"crypto/rand"
	"encoding/hex"
)

func generateSecureToken(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func GenerateTokenAndCrf() (string, string, error) {
	csrf, err := generateSecureToken(16)
	if err != nil {
		return "", "", err
	}
	token, err := generateSecureToken(16)
	if err != nil {
		return "", "", err
	}
	return csrf, token, nil
}
