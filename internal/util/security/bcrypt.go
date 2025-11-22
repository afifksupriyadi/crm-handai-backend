package security

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

// defaultCost defines the computational cost for hashing passwords.
// Higher values increase security but slow down processing.
const defaultCost = 12

var (
	// ErrEmptyPassword is returned when attempting to hash an empty password.
	ErrEmptyPassword = errors.New("password cannot be empty")
)

// HashPassword hashes the given plaintext password using bcrypt.
// It returns the resulting hash or an error if the password is invalid.
func HashPassword(password string) (string, error) {
	if password == "" {
		return "", ErrEmptyPassword
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), defaultCost)
	if err != nil {
		return "", err
	}

	return string(hashed), nil
}

// ComparePassword compares a plaintext password against a bcrypt hash.
// It returns true if the password matches the hash, otherwise false.
func ComparePassword(password, hashed string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))
	return err == nil
}
