package services

import (
	"auth-service/internal/application/ports"
	"auth-service/internal/domain"
	"errors"
)

// Allgemeine Fehlerdefinitionen fuer Authentifizierungs-Operationen.
var (
	ErrUserAlreadyExists  = errors.New("Benutzer mit dieser E-Mail existiert bereits")
	ErrInvalidCredentials = errors.New("Ungueltige E-Mail oder Passwort")
	ErrUserNotFound       = errors.New("Benutzer nicht gefunden")
	ErrInvalidRole        = errors.New("Ungueltige Rolle angegeben")
)

// AuthServiceImpl implementiert das AuthService-Interface.
// Es koordiniert den Authentifizierungs-Workflow mit injizierten Abhaengigkeiten.
type AuthServiceImpl struct {
	userRepo       ports.UserRepository
	passwordHasher ports.PasswordHasher
	tokenGenerator ports.TokenGenerator
}

// NewAuthService erstellt eine neue AuthService-Instanz mit den erforderlichen Abhaengigkeiten.
// Dies folgt dem Dependency-Injection-Pattern fuer bessere Testbarkeit.
func NewAuthService(
	userRepo ports.UserRepository,
	passwordHasher ports.PasswordHasher,
	tokenGenerator ports.TokenGenerator,
) *AuthServiceImpl {
	return &AuthServiceImpl{
		userRepo:       userRepo,
		passwordHasher: passwordHasher,
		tokenGenerator: tokenGenerator,
	}
}

// Register erstellt ein neues Benutzerkonto nach Validierung der Eingabe.
// Das Passwort wird vor dem Speichern gehasht um Sicherheit zu gewaehrleisten.
func (s *AuthServiceImpl) Register(email, password, role string) error {
	// Rolle vor dem Fortfahren validieren
	if !domain.ValidateRole(role) {
		return ErrInvalidRole
	}

	// Pruefen ob Benutzer bereits existiert
	exists, err := s.userRepo.ExistsByEmail(email)
	if err != nil {
		return err
	}
	if exists {
		return ErrUserAlreadyExists
	}

	// Passwort fuer sichere Speicherung hashen
	hashedPassword, err := s.passwordHasher.Hash(password)
	if err != nil {
		return err
	}

	// Neuen Benutzer erstellen und persistieren
	user := domain.NewUser(email, hashedPassword, role)
	return s.userRepo.Create(user)
}

// Login authentifiziert den Benutzer und generiert bei Erfolg ein JWT-Token.
func (s *AuthServiceImpl) Login(email, password string) (string, error) {
	user, err := s.Authenticate(email, password)
	if err != nil {
		return "", err
	}

	return s.tokenGenerator.Generate(user)
}

// Authenticate verifiziert Benutzeranmeldedaten ohne Token-Generierung.
// Gibt die Benutzer-Entitaet zurueck wenn die Authentifizierung erfolgreich ist.
func (s *AuthServiceImpl) Authenticate(email, password string) (*domain.User, error) {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if !s.passwordHasher.Compare(user.Password, password) {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}

// GetUserByID ruft einen Benutzer anhand seiner eindeutigen ID ab.
func (s *AuthServiceImpl) GetUserByID(id string) (*domain.User, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}
