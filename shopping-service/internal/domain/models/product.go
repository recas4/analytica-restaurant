package domain

import "go.mongodb.org/mongo-driver/bson/primitive"

// Product repraesentiert ein Produkt im Shopping-System.
// Enthaelt Produktinformationen wie Name, Preis und den Ersteller.
type Product struct {
	ID     primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name   string             `json:"name" bson:"name"`
	Price  float64            `json:"price" bson:"price"`
	UserID string             `json:"user_id" bson:"user_id"`
}

// NewProduct erstellt ein neues Produkt mit den angegebenen Werten.
func NewProduct(name string, price float64, userID string) *Product {
	return &Product{
		Name:   name,
		Price:  price,
		UserID: userID,
	}
}

// IsValid prueft ob das Produkt gueltige Werte hat.
func (p *Product) IsValid() bool {
	return p.Name != "" && p.Price >= 0
}
