package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
)

// User represents a user in the system
type User struct {
	ID           int
	Username     string
	PasswordHash string
	Email        string
	CreatedAt    time.Time
	IsActive     bool
}

// AuthService handles authentication
type AuthService struct {
	users map[string]*User
}

// NewAuthService creates a new auth service
func NewAuthService() *AuthService {
	return &AuthService{
		users: make(map[string]*User),
	}
}

// Register creates a new user account
func (s *AuthService) Register(username, password, email string) (*User, error) {
	if err := ValidateUsername(username); err != nil {
		return nil, err
	}
	if err := ValidatePassword(password); err != nil {
		return nil, err
	}
	if err := ValidateEmail(email); err != nil {
		return nil, err
	}

	if _, exists := s.users[username]; exists {
		return nil, errors.New("username already exists")
	}

	user := &User{
		ID:           len(s.users) + 1,
		Username:     username,
		PasswordHash: HashPassword(password),
		Email:        email,
		CreatedAt:    time.Now(),
		IsActive:     true,
	}

	s.users[username] = user
	return user, nil
}

// Login authenticates a user
func (s *AuthService) Login(username, password string) (*User, error) {
	user, exists := s.users[username]
	if !exists {
		return nil, errors.New("invalid credentials")
	}

	if !user.IsActive {
		return nil, errors.New("account is inactive")
	}

	if !VerifyPassword(password, user.PasswordHash) {
		return nil, errors.New("invalid credentials")
	}

	return user, nil
}

// DeactivateUser deactivates a user account
func (s *AuthService) DeactivateUser(username string) error {
	user, exists := s.users[username]
	if !exists {
		return fmt.Errorf("user %s not found", username)
	}

	user.IsActive = false
	return nil
}

// ValidateUsername checks if username is valid
func ValidateUsername(username string) error {
	if len(username) < 3 {
		return errors.New("username must be at least 3 characters")
	}
	if len(username) > 20 {
		return errors.New("username must be at most 20 characters")
	}
	for _, char := range username {
		if !isAlphanumeric(char) {
			return errors.New("username must contain only alphanumeric characters")
		}
	}
	return nil
}

// ValidatePassword checks if password is valid
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	if len(password) > 50 {
		return errors.New("password must be at most 50 characters")
	}
	return nil
}

// ValidateEmail checks if email is valid (simple validation)
func ValidateEmail(email string) error {
	if len(email) < 5 {
		return errors.New("email is too short")
	}
	if !strings.Contains(email, "@") {
		return errors.New("email must contain @")
	}
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return errors.New("invalid email format")
	}
	if len(parts[0]) == 0 || len(parts[1]) == 0 {
		return errors.New("invalid email format")
	}
	return nil
}

// HashPassword creates a hash from password
func HashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

// VerifyPassword checks if password matches hash
func VerifyPassword(password, hash string) bool {
	return HashPassword(password) == hash
}

// isAlphanumeric checks if character is alphanumeric
func isAlphanumeric(char rune) bool {
	return (char >= 'a' && char <= 'z') ||
		(char >= 'A' && char <= 'Z') ||
		(char >= '0' && char <= '9')
}
