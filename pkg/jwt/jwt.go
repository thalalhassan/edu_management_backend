package jwt

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	refreshTokenBytes = 32 // 256-bit random token
	RefreshTokenTTL   = 7 * 24 * time.Hour
	DefaultAccessTTL  = 15 * time.Minute
)

// Claims is the JWT payload embedded in every access token.
type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateAccessToken signs a new JWT access token.
// ttl controls expiry — pass crypto.DefaultAccessTTL for the default.
func GenerateAccessToken(userID, role, secret string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("crypto.GenerateAccessToken: %w", err)
	}
	return signed, nil
}

// ParseAccessToken validates a signed JWT and returns its claims.
// Returns an error if the token is expired, malformed, or the signature is invalid.
func ParseAccessToken(tokenStr, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("crypto.ParseAccessToken: unexpected signing method %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("crypto.ParseAccessToken: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("crypto.ParseAccessToken: invalid token")
	}
	return claims, nil
}

// GenerateRefreshToken produces a cryptographically random opaque token
// and its expiry time. The raw string is what gets stored in the DB and
// sent to the client — it is NOT a JWT.
func GenerateRefreshToken() (raw string, expiresAt time.Time, err error) {
	b := make([]byte, refreshTokenBytes)
	if _, err = rand.Read(b); err != nil {
		return "", time.Time{}, fmt.Errorf("crypto.GenerateRefreshToken: %w", err)
	}
	return hex.EncodeToString(b), time.Now().Add(RefreshTokenTTL), nil
}
