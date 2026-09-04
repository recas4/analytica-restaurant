package ports

import "shopping-service/internal/domain/models"

// ProductService definiert die Geschaeftslogik fuer Produkte.
// Wird vom ProductServiceImpl implementiert.
type ProductService interface {
	// CreateProduct erstellt ein neues Produkt.
	CreateProduct(product *domain.Product) error

	// GetAllProducts liefert alle verfuegbaren Produkte.
	GetAllProducts() ([]domain.Product, error)

	// GetProductByID liefert ein Produkt anhand seiner ID.
	GetProductByID(id string) (*domain.Product, error)
}
