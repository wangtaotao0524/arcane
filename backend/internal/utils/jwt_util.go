package utils

import (
	"crypto/rand"
	"fmt"
)

func CheckOrGenerateJwtSecret(jwtSecret string) []byte {
	var secretBytes []byte
	if jwtSecret != "" {
		secretBytes = []byte(jwtSecret)
		return secretBytes
	} else {
		secretBytes = make([]byte, 32)
		if _, err := rand.Read(secretBytes); err != nil {
			panic(fmt.Errorf("failed to generate random JWT secret: %w", err))
		}
	}
	return nil
}
