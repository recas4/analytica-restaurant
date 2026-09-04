package ports

import "auth-service/internal/domain"

// TokenGenerator definiert das Interface fuer JWT-Token-Operationen.
// Diese Abstraktion entkoppelt die Token-Generierung von der Geschaeftslogik.
type TokenGenerator interface {
	// Generate erstellt ein neues JWT-Token fuer den angegebenen Benutzer.
	// Gibt den signierten Token-String oder einen Fehler zurueck wenn die Generierung fehlschlaegt.
	Generate(user *domain.User) (string, error)

	// Validate prueft das Token und gibt die User-ID zurueck wenn es gueltig ist.
	// Gibt einen Fehler zurueck wenn das Token ungueltig oder abgelaufen ist.
	Validate(token string) (string, error)
}
