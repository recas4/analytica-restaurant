package services

import (
	"shopping-service/internal/application/ports"
	"shopping-service/internal/domain/models"
)

// CartServiceImpl implementiert das CartService-Interface.
// Koordiniert Warenkorb-Operationen zwischen Handler und Repository.
type CartServiceImpl struct {
	repo ports.CartRepository
}

// NewCartService erstellt eine neue CartService-Instanz.
func NewCartService(repo ports.CartRepository) ports.CartService {
	return &CartServiceImpl{repo: repo}
}

// AddToCart fuegt einen Artikel zum Warenkorb hinzu.
func (s *CartServiceImpl) AddToCart(userID, productID string, qty int) error {
	return s.repo.AddItem(userID, productID, qty)
}

// GetCart liefert den aktuellen Warenkorb des Benutzers.
func (s *CartServiceImpl) GetCart(userID string) (domain.Cart, error) {
	return s.repo.GetCart(userID)
}

// ClearCart leert den Warenkorb des Benutzers.
func (s *CartServiceImpl) ClearCart(userID string) error {
	return s.repo.ClearCart(userID)
}

// UpdateCartItem aktualisiert die Menge eines Artikels im Warenkorb.
func (s *CartServiceImpl) UpdateCartItem(userID, productID string, qty int) error {
	return s.repo.UpdateItem(userID, productID, qty)
}

// RemoveFromCart entfernt einen Artikel aus dem Warenkorb.
func (s *CartServiceImpl) RemoveFromCart(userID, productID string) error {
	return s.repo.RemoveItem(userID, productID)
}
