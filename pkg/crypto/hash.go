package crypto

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const defaultCost = bcrypt.DefaultCost

// Hash generates a bcrypt hash from a plain-text password.
func Hash(plain string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(plain), defaultCost)
	if err != nil {
		return "", fmt.Errorf("crypto.Hash: %w", err)
	}
	return string(bytes), nil
}

// CheckHash compares a plain-text password against a bcrypt hash.
// Returns true if they match.
func CheckHash(plain, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}
