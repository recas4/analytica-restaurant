package services

import (
	"shopping-service/internal/application/ports"
	"shopping-service/internal/domain/models"
)

// ProductServiceImpl implementiert das ProductService-Interface.
// Koordiniert Produkt-Operationen zwischen Handler und Repository.
type ProductServiceImpl struct {
	repo ports.ProductRepository
}

// NewProductService erstellt eine neue ProductService-Instanz.
func NewProductService(repo ports.ProductRepository) ports.ProductService {
	return &ProductServiceImpl{repo: repo}
}

// CreateProduct speichert ein neues Produkt.
func (s *ProductServiceImpl) CreateProduct(product *domain.Product) error {
	return s.repo.Create(product)
}

// GetAllProducts liefert alle verfuegbaren Produkte.
func (s *ProductServiceImpl) GetAllProducts() ([]domain.Product, error) {
	return s.repo.FindAll()
}

// GetProductByID sucht ein Produkt anhand seiner ID.
func (s *ProductServiceImpl) GetProductByID(id string) (*domain.Product, error) {
	return s.repo.FindByID(id)
}
