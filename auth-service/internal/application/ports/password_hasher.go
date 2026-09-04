package ports

// PasswordHasher definiert das Interface fuer Passwort-Hashing-Operationen.
// Diese Abstraktion ermoeglicht den Austausch von Hashing-Algorithmen ohne Aenderung der Geschaeftslogik.
type PasswordHasher interface {
	// Hash generiert einen sicheren Hash aus dem Klartext-Passwort.
	// Gibt das gehashte Passwort oder einen Fehler zurueck wenn das Hashing fehlschlaegt.
	Hash(password string) (string, error)

	// Compare prueft ob das Klartext-Passwort mit dem gehashten Passwort uebereinstimmt.
	// Gibt true zurueck wenn sie uebereinstimmen, sonst false.
	Compare(hashedPassword, plainPassword string) bool
}
