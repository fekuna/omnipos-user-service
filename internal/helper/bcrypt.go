package helper

import (
	"golang.org/x/crypto/bcrypt"
)

const (
	// DefaultCost is the default bcrypt cost factor (10)
	// Higher values = more secure but slower
	// Range: 4-31, recommended: 10-14
	DefaultCost = bcrypt.DefaultCost
)

// HashPassword generates a bcrypt hash from a plain text password
// Returns the hashed password or an error if hashing fails
func HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// HashPasswordWithCost generates a bcrypt hash with a custom cost factor
// Use this when you need stronger/weaker hashing than the default
func HashPasswordWithCost(password string, cost int) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// ComparePassword compares a hashed password with a plain text password
// Returns true if they match, false otherwise
func ComparePassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// ValidatePasswordStrength checks if a password meets minimum requirements
// Returns true if password is acceptable
func ValidatePasswordStrength(password string) bool {
	// Minimum 4 digits for PIN
	// Adjust this based on your business requirements
	return len(password) >= 4
}
