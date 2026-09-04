package bcrypt

import (
	"auth-service/internal/application/ports"

	"golang.org/x/crypto/bcrypt"
)

// BcryptPasswordHasher implementiert PasswordHasher mit dem bcrypt-Algorithmus.
// Bcrypt wird fuer Passwort-Hashing empfohlen aufgrund seines adaptiven Kostenfaktors.
type BcryptPasswordHasher struct {
	cost int
}

// NewBcryptPasswordHasher erstellt einen neuen Hasher mit dem Standard-bcrypt-Kostenfaktor.
func NewBcryptPasswordHasher() ports.PasswordHasher {
	return &BcryptPasswordHasher{
		cost: bcrypt.DefaultCost,
	}
}

// NewBcryptPasswordHasherWithCost erstellt einen Hasher mit einem benutzerdefinierten Kostenfaktor.
// Hoehere Kosten erhoehen die Sicherheit aber auch die Verarbeitungszeit.
func NewBcryptPasswordHasherWithCost(cost int) ports.PasswordHasher {
	return &BcryptPasswordHasher{
		cost: cost,
	}
}

// Hash generiert einen bcrypt-Hash aus dem Klartext-Passwort.
func (h *BcryptPasswordHasher) Hash(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// Compare prueft ob das Klartext-Passwort mit dem gehashten Passwort uebereinstimmt.
func (h *BcryptPasswordHasher) Compare(hashedPassword, plainPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	return err == nil
}
