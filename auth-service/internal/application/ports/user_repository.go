package ports

import "auth-service/internal/domain"

// UserRepository definiert das Interface fuer Benutzer-Persistenz-Operationen.
// Dieser Port abstrahiert die Datenschicht von der Geschaeftslogik,
// und ermoeglicht verschiedene Speicher-Implementierungen (MongoDB, PostgreSQL, etc.).
type UserRepository interface {
	// Create speichert einen neuen Benutzer in der Datenbank.
	// Gibt einen Fehler zurueck wenn die Operation fehlschlaegt.
	Create(user *domain.User) error

	// FindByEmail ruft einen Benutzer anhand seiner E-Mail-Adresse ab.
	// Gibt nil und einen Fehler zurueck wenn der Benutzer nicht gefunden wird.
	FindByEmail(email string) (*domain.User, error)

	// FindByID ruft einen Benutzer anhand seiner eindeutigen ID ab.
	// Gibt nil und einen Fehler zurueck wenn der Benutzer nicht gefunden wird.
	FindByID(id string) (*domain.User, error)

	// ExistsByEmail prueft ob ein Benutzer mit der angegebenen E-Mail bereits existiert.
	ExistsByEmail(email string) (bool, error)
}
