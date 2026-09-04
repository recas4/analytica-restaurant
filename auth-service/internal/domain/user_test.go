package domain

import "testing"

func TestNewUser(t *testing.T) {
	user := NewUser("test@example.com", "hashed_password", "user")

	if user == nil {
		t.Error("NewUser should not return nil")
	}

	if user.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got '%s'", user.Email)
	}

	if user.Password != "hashed_password" {
		t.Errorf("Expected password 'hashed_password', got '%s'", user.Password)
	}

	if user.Role != "user" {
		t.Errorf("Expected role 'user', got '%s'", user.Role)
	}
}

func TestUser_IsAdmin_True(t *testing.T) {
	user := NewUser("admin@example.com", "password", "admin")

	if !user.IsAdmin() {
		t.Error("User with role 'admin' should return true for IsAdmin()")
	}
}

func TestUser_IsAdmin_False(t *testing.T) {
	user := NewUser("user@example.com", "password", "user")

	if user.IsAdmin() {
		t.Error("User with role 'user' should return false for IsAdmin()")
	}
}

func TestValidateRole_ValidRoles(t *testing.T) {
	testCases := []struct {
		role     string
		expected bool
	}{
		{"admin", true},
		{"user", true},
		{"Admin", false},
		{"USER", false},
		{"guest", false},
		{"", false},
		{"superadmin", false},
	}

	for _, tc := range testCases {
		result := ValidateRole(tc.role)
		if result != tc.expected {
			t.Errorf("ValidateRole('%s') = %v, expected %v", tc.role, result, tc.expected)
		}
	}
}
