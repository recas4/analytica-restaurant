package ports

import "shopping-service/internal/domain/models"

// CartRepository definiert die Schnittstelle fuer Warenkorb-Persistenz.
// Wird von der MongoDB-Implementierung umgesetzt.
type CartRepository interface {
	// AddItem fuegt einen Artikel zum Warenkorb hinzu.
	AddItem(userID, productID string, qty int) error

	// GetCart liefert den Warenkorb eines Benutzers.
	GetCart(userID string) (domain.Cart, error)

	// ClearCart leert den Warenkorb eines Benutzers.
	ClearCart(userID string) error

	// UpdateItem aktualisiert die Menge eines Artikels.
	UpdateItem(userID, productID string, qty int) error

	// RemoveItem entfernt einen Artikel aus dem Warenkorb.
	RemoveItem(userID, productID string) error
}
