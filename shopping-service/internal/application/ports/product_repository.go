package ports

import "shopping-service/internal/domain/models"

// ProductRepository definiert die Schnittstelle fuer Produkt-Persistenz.
// Wird von der MongoDB-Implementierung umgesetzt.
type ProductRepository interface {
	// Create speichert ein neues Produkt in der Datenbank.
	Create(product *domain.Product) error

	// FindAll liefert alle Produkte aus der Datenbank.
	FindAll() ([]domain.Product, error)

	// FindByID sucht ein Produkt anhand seiner ID.
	FindByID(id string) (*domain.Product, error)
}
