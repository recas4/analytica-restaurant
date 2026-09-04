package ports

import "shopping-service/internal/domain/models"

// CartService definiert die Geschaeftslogik fuer den Warenkorb.
// Wird vom CartServiceImpl implementiert.
type CartService interface {
	// AddToCart fuegt einen Artikel zum Warenkorb hinzu.
	AddToCart(userID, productID string, qty int) error

	// GetCart liefert den aktuellen Warenkorb.
	GetCart(userID string) (domain.Cart, error)

	// ClearCart leert den Warenkorb.
	ClearCart(userID string) error

	// UpdateCartItem aktualisiert die Menge eines Artikels.
	UpdateCartItem(userID, productID string, qty int) error

	// RemoveFromCart entfernt einen Artikel aus dem Warenkorb.
	RemoveFromCart(userID, productID string) error
}
