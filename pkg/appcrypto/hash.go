package appcrypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"
)

const defaultCost = bcrypt.DefaultCost

// BcryptHash generates a bcrypt hash from a plain-text password.
func BcryptHash(plain string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(plain), defaultCost)
	if err != nil {
		return "", fmt.Errorf("appcrypto.BcryptHash: %w", err)
	}
	return string(bytes), nil
}

// BcryptVerifyHash compares a plain-text password against a bcrypt hash.
// Returns true if they match.
func BcryptVerifyHash(plain, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}

func Hash(raw string) string {
	var tokenSecret = []byte(os.Getenv("HASH_SECRET"))
	mac := hmac.New(sha256.New, tokenSecret)
	mac.Write([]byte(raw))
	return hex.EncodeToString(mac.Sum(nil))
}

func VerifyHash(rawToken, storedHash string) bool {
	expected := Hash(rawToken)
	return hmac.Equal([]byte(expected), []byte(storedHash))
}
