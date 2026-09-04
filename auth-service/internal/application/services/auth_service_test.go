package services

import (
	"auth-service/internal/domain"
	"errors"
	"testing"
)

// Mock implementations for testing

type mockUserRepository struct {
	users       map[string]*domain.User
	shouldError bool
}

func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{
		users: make(map[string]*domain.User),
	}
}

func (m *mockUserRepository) Create(user *domain.User) error {
	if m.shouldError {
		return errors.New("database error")
	}
	user.ID = "generated-id-123"
	m.users[user.Email] = user
	return nil
}

func (m *mockUserRepository) FindByEmail(email string) (*domain.User, error) {
	if m.shouldError {
		return nil, errors.New("database error")
	}
	user, exists := m.users[email]
	if !exists {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (m *mockUserRepository) FindByID(id string) (*domain.User, error) {
	if m.shouldError {
		return nil, errors.New("database error")
	}
	for _, user := range m.users {
		if user.ID == id {
			return user, nil
		}
	}
	return nil, errors.New("user not found")
}

func (m *mockUserRepository) ExistsByEmail(email string) (bool, error) {
	if m.shouldError {
		return false, errors.New("database error")
	}
	_, exists := m.users[email]
	return exists, nil
}

type mockPasswordHasher struct {
	shouldError bool
}

func (m *mockPasswordHasher) Hash(password string) (string, error) {
	if m.shouldError {
		return "", errors.New("hash error")
	}
	return "hashed_" + password, nil
}

func (m *mockPasswordHasher) Compare(hashedPassword, password string) bool {
	return hashedPassword == "hashed_"+password
}

type mockTokenGenerator struct {
	shouldError bool
}

func (m *mockTokenGenerator) Generate(user *domain.User) (string, error) {
	if m.shouldError {
		return "", errors.New("token generation error")
	}
	return "mock_token_for_" + user.Email, nil
}

func (m *mockTokenGenerator) Validate(token string) (string, error) {
	if m.shouldError {
		return "", errors.New("token validation error")
	}
	return "user-id-123", nil
}

// Tests

func TestNewAuthService(t *testing.T) {
	repo := newMockUserRepository()
	hasher := &mockPasswordHasher{}
	tokenGen := &mockTokenGenerator{}

	service := NewAuthService(repo, hasher, tokenGen)

	if service == nil {
		t.Error("NewAuthService should not return nil")
	}
}

func TestRegister_Success(t *testing.T) {
	repo := newMockUserRepository()
	hasher := &mockPasswordHasher{}
	tokenGen := &mockTokenGenerator{}
	service := NewAuthService(repo, hasher, tokenGen)

	err := service.Register("test@example.com", "password123", "user")

	if err != nil {
		t.Errorf("Register should succeed, got error: %v", err)
	}

	if len(repo.users) != 1 {
		t.Errorf("Expected 1 user in repository, got %d", len(repo.users))
	}
}

func TestRegister_InvalidRole(t *testing.T) {
	repo := newMockUserRepository()
	hasher := &mockPasswordHasher{}
	tokenGen := &mockTokenGenerator{}
	service := NewAuthService(repo, hasher, tokenGen)

	err := service.Register("test@example.com", "password123", "invalid_role")

	if err != ErrInvalidRole {
		t.Errorf("Expected ErrInvalidRole, got: %v", err)
	}
}

func TestRegister_UserAlreadyExists(t *testing.T) {
	repo := newMockUserRepository()
	hasher := &mockPasswordHasher{}
	tokenGen := &mockTokenGenerator{}
	service := NewAuthService(repo, hasher, tokenGen)

	// erste Registrierung
	_ = service.Register("test@example.com", "password123", "user")

	// zweite Registrierung mit derselben E-Mail
	err := service.Register("test@example.com", "different_password", "user")

	if err != ErrUserAlreadyExists {
		t.Errorf("Expected ErrUserAlreadyExists, got: %v", err)
	}
}

func TestRegister_AdminRole(t *testing.T) {
	repo := newMockUserRepository()
	hasher := &mockPasswordHasher{}
	tokenGen := &mockTokenGenerator{}
	service := NewAuthService(repo, hasher, tokenGen)

	err := service.Register("admin@example.com", "admin123", "admin")

	if err != nil {
		t.Errorf("Register with admin role should succeed, got error: %v", err)
	}

	user := repo.users["admin@example.com"]
	if user.Role != "admin" {
		t.Errorf("Expected role 'admin', got '%s'", user.Role)
	}
}

func TestLogin_Success(t *testing.T) {
	repo := newMockUserRepository()
	hasher := &mockPasswordHasher{}
	tokenGen := &mockTokenGenerator{}
	service := NewAuthService(repo, hasher, tokenGen)

	// Register user erst
	_ = service.Register("test@example.com", "password123", "user")

	// Login
	token, err := service.Login("test@example.com", "password123")

	if err != nil {
		t.Errorf("Login should succeed, got error: %v", err)
	}

	if token == "" {
		t.Error("Token should not be empty")
	}

	expectedToken := "mock_token_for_test@example.com"
	if token != expectedToken {
		t.Errorf("Expected token '%s', got '%s'", expectedToken, token)
	}
}

func TestLogin_InvalidPassword(t *testing.T) {
	repo := newMockUserRepository()
	hasher := &mockPasswordHasher{}
	tokenGen := &mockTokenGenerator{}
	service := NewAuthService(repo, hasher, tokenGen)

	// Register user erst
	_ = service.Register("test@example.com", "password123", "user")

	// Login mit falschem Passwort
	_, err := service.Login("test@example.com", "wrong_password")

	if err != ErrInvalidCredentials {
		t.Errorf("Expected ErrInvalidCredentials, got: %v", err)
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	repo := newMockUserRepository()
	hasher := &mockPasswordHasher{}
	tokenGen := &mockTokenGenerator{}
	service := NewAuthService(repo, hasher, tokenGen)

	_, err := service.Login("nonexistent@example.com", "password123")

	if err != ErrInvalidCredentials {
		t.Errorf("Expected ErrInvalidCredentials, got: %v", err)
	}
}

func TestAuthenticate_Success(t *testing.T) {
	repo := newMockUserRepository()
	hasher := &mockPasswordHasher{}
	tokenGen := &mockTokenGenerator{}
	service := NewAuthService(repo, hasher, tokenGen)

	_ = service.Register("test@example.com", "password123", "user")

	user, err := service.Authenticate("test@example.com", "password123")

	if err != nil {
		t.Errorf("Authenticate should succeed, got error: %v", err)
	}

	if user == nil {
		t.Error("User should not be nil")
	}

	if user.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got '%s'", user.Email)
	}
}

func TestGetUserByID_Success(t *testing.T) {
	repo := newMockUserRepository()
	hasher := &mockPasswordHasher{}
	tokenGen := &mockTokenGenerator{}
	service := NewAuthService(repo, hasher, tokenGen)

	_ = service.Register("test@example.com", "password123", "user")

	user, err := service.GetUserByID("generated-id-123")

	if err != nil {
		t.Errorf("GetUserByID should succeed, got error: %v", err)
	}

	if user == nil {
		t.Error("User should not be nil")
	}
}

func TestGetUserByID_NotFound(t *testing.T) {
	repo := newMockUserRepository()
	hasher := &mockPasswordHasher{}
	tokenGen := &mockTokenGenerator{}
	service := NewAuthService(repo, hasher, tokenGen)

	_, err := service.GetUserByID("nonexistent-id")

	if err != ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got: %v", err)
	}
}

func TestRegister_HashError(t *testing.T) {
	repo := newMockUserRepository()
	hasher := &mockPasswordHasher{shouldError: true}
	tokenGen := &mockTokenGenerator{}
	service := NewAuthService(repo, hasher, tokenGen)

	err := service.Register("test@example.com", "password123", "user")

	if err == nil {
		t.Error("Expected error when hasher fails")
	}
}

func TestLogin_TokenGenerationError(t *testing.T) {
	repo := newMockUserRepository()
	hasher := &mockPasswordHasher{}
	tokenGen := &mockTokenGenerator{shouldError: true}
	service := NewAuthService(repo, hasher, tokenGen)

	_ = service.Register("test@example.com", "password123", "user")

	_, err := service.Login("test@example.com", "password123")

	if err == nil {
		t.Error("Expected error when token generation fails")
	}
}
