package ports

import "auth-service/internal/domain"

// AuthService definiert das Interface fuer Authentifizierungs-Operationen.
// Dies ist der primaere Port den die Anwendung externen Adaptern bereitstellt.
type AuthService interface {
	// Register erstellt ein neues Benutzerkonto mit den angegebenen Anmeldedaten.
	// Gibt einen Fehler zurueck wenn die E-Mail bereits vergeben ist oder die Validierung fehlschlaegt.
	Register(email, password, role string) error

	// Login authentifiziert einen Benutzer und gibt ein JWT-Token zurueck.
	// Gibt einen Fehler zurueck wenn die Anmeldedaten ungueltig sind.
	Login(email, password string) (string, error)

	// Authenticate validiert Anmeldedaten und gibt den Benutzer ohne Token-Generierung zurueck.
	// Nuetzlich fuer interne Authentifizierungspruefungen.
	Authenticate(email, password string) (*domain.User, error)

	// GetUserByID ruft Benutzerinformationen anhand der ID ab.
	// Gibt einen Fehler zurueck wenn der Benutzer nicht gefunden wird.
	GetUserByID(id string) (*domain.User, error)
}
