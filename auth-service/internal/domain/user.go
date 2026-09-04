package domain

// User repraesentiert die zentrale Benutzer-Entitaet in der Authentifizierungs-Domain.
// Dies ist das zentrale Geschaeftsobjekt das Benutzeridentitaetsinformationen haelt.
type User struct {
	ID       string `bson:"_id,omitempty" json:"id"`
	Email    string `bson:"email" json:"email"`
	Password string `bson:"password" json:"-"`
	Role     string `bson:"role" json:"role"`
}

// NewUser erstellt eine neue User-Instanz mit den angegebenen Attributen.
// Das Passwort sollte bereits gehasht sein bevor diese Funktion aufgerufen wird.
func NewUser(email, hashedPassword, role string) *User {
	return &User{
		Email:    email,
		Password: hashedPassword,
		Role:     role,
	}
}

// IsAdmin prueft ob der Benutzer Administratorrechte hat.
func (u *User) IsAdmin() bool {
	return u.Role == "admin"
}

// ValidateRole stellt sicher dass die Rolle entweder "admin" oder "user" ist.
// Gibt true zurueck wenn die Rolle gueltig ist, sonst false.
func ValidateRole(role string) bool {
	return role == "admin" || role == "user"
}
