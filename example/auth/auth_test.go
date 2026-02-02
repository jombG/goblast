package auth

import (
	"testing"
)

func TestNewAuthService(t *testing.T) {
	service := NewAuthService()
	if service == nil {
		t.Fatal("expected non-nil service")
	}
}

func TestRegister(t *testing.T) {
	service := NewAuthService()

	tests := []struct {
		name      string
		username  string
		password  string
		email     string
		wantError bool
	}{
		{"valid user", "john", "password123", "john@example.com", false},
		{"short username", "ab", "password123", "test@example.com", true},
		{"long username", "verylongusername12345", "password123", "test@example.com", true},
		{"short password", "alice", "pass", "alice@example.com", true},
		{"invalid email", "bob", "password123", "invalid", true},
		{"valid user 2", "alice", "securepass", "alice@test.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := service.Register(tt.username, tt.password, tt.email)
			if tt.wantError {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if user == nil {
					t.Error("expected user, got nil")
				}
				if user.Username != tt.username {
					t.Errorf("expected username %s, got %s", tt.username, user.Username)
				}
				if !user.IsActive {
					t.Error("expected user to be active")
				}
			}
		})
	}

	// Test duplicate username
	_, err := service.Register("john", "password456", "john2@example.com")
	if err == nil {
		t.Error("expected error for duplicate username")
	}
}

func TestLogin(t *testing.T) {
	service := NewAuthService()
	service.Register("testuser", "password123", "test@example.com")

	// Valid login
	user, err := service.Login("testuser", "password123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Username != "testuser" {
		t.Errorf("expected username testuser, got %s", user.Username)
	}

	// Invalid password
	_, err = service.Login("testuser", "wrongpassword")
	if err == nil {
		t.Error("expected error for wrong password")
	}

	// Non-existing user
	_, err = service.Login("nonexistent", "password123")
	if err == nil {
		t.Error("expected error for non-existing user")
	}

	// Deactivated user
	service.DeactivateUser("testuser")
	_, err = service.Login("testuser", "password123")
	if err == nil {
		t.Error("expected error for deactivated user")
	}
}

func TestDeactivateUser(t *testing.T) {
	service := NewAuthService()
	service.Register("testuser", "password123", "test@example.com")

	// Deactivate existing user
	err := service.DeactivateUser("testuser")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Try to deactivate non-existing user
	err = service.DeactivateUser("nonexistent")
	if err == nil {
		t.Error("expected error for non-existing user")
	}
}

func TestValidateUsername(t *testing.T) {
	tests := []struct {
		username  string
		wantError bool
	}{
		{"john", false},
		{"user123", false},
		{"ab", true},
		{"verylongusername12345", true},
		{"user@name", true},
		{"user name", true},
	}

	for _, tt := range tests {
		err := ValidateUsername(tt.username)
		if tt.wantError && err == nil {
			t.Errorf("ValidateUsername(%s) expected error, got nil", tt.username)
		}
		if !tt.wantError && err != nil {
			t.Errorf("ValidateUsername(%s) unexpected error: %v", tt.username, err)
		}
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		password  string
		wantError bool
	}{
		{"password123", false},
		{"short", true},
		{"verylongpasswordthatexceedsfiftycharacterslimitforusersecurity", true},
		{"12345678", false},
	}

	for _, tt := range tests {
		err := ValidatePassword(tt.password)
		if tt.wantError && err == nil {
			t.Errorf("ValidatePassword(%s) expected error, got nil", tt.password)
		}
		if !tt.wantError && err != nil {
			t.Errorf("ValidatePassword(%s) unexpected error: %v", tt.password, err)
		}
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		email     string
		wantError bool
	}{
		{"user@example.com", false},
		{"test@test.co", false},
		{"invalid", true},
		{"@example.com", true},
		{"user@", true},
		{"a@b", false},
		{"user@@example.com", true},
	}

	for _, tt := range tests {
		err := ValidateEmail(tt.email)
		if tt.wantError && err == nil {
			t.Errorf("ValidateEmail(%s) expected error, got nil", tt.email)
		}
		if !tt.wantError && err != nil {
			t.Errorf("ValidateEmail(%s) unexpected error: %v", tt.email, err)
		}
	}
}

func TestHashPassword(t *testing.T) {
	password := "testpassword123"
	hash1 := HashPassword(password)
	hash2 := HashPassword(password)

	// Same password should produce same hash
	if hash1 != hash2 {
		t.Error("same password produced different hashes")
	}

	// Different password should produce different hash
	hash3 := HashPassword("differentpassword")
	if hash1 == hash3 {
		t.Error("different passwords produced same hash")
	}

	// Hash should be non-empty
	if hash1 == "" {
		t.Error("hash is empty")
	}
}

func TestVerifyPassword(t *testing.T) {
	password := "testpassword123"
	hash := HashPassword(password)

	// Correct password
	if !VerifyPassword(password, hash) {
		t.Error("failed to verify correct password")
	}

	// Wrong password
	if VerifyPassword("wrongpassword", hash) {
		t.Error("verified wrong password")
	}
}
