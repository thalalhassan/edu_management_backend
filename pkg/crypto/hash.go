package crypto

import "golang.org/x/crypto/bcrypt"

// Helper functions

// hashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	const cost = bcrypt.DefaultCost
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

// comparePasswords compares a hashed password with a plain text password
func ComparePasswords(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
