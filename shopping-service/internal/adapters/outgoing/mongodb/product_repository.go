package mongodb

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"shopping-service/internal/application/ports"
	"shopping-service/internal/domain/models"
)

// ProductRepository implementiert das ProductRepository-Interface mit MongoDB.
type ProductRepository struct {
	collection *mongo.Collection
}

// NewProductRepository erstellt eine neue ProductRepository-Instanz.
func NewProductRepository(db *mongo.Database) ports.ProductRepository {
	return &ProductRepository{
		collection: db.Collection("products"),
	}
}

// Create speichert ein neues Produkt in der Datenbank.
func (r *ProductRepository) Create(product *domain.Product) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := r.collection.InsertOne(ctx, product)
	if err != nil {
		log.Printf("Fehler beim Einfuegen: %v", err)
		return err
	}

	log.Printf("Produkt eingefuegt mit ID: %v", result.InsertedID)
	return nil
}

// FindAll liefert alle Produkte aus der Datenbank.
func (r *ProductRepository) FindAll() ([]domain.Product, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		log.Printf("Fehler bei Find: %v", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var products []domain.Product
	for cursor.Next(ctx) {
		var p domain.Product
		if err := cursor.Decode(&p); err != nil {
			log.Printf("Fehler beim Dekodieren: %v", err)
			return nil, err
		}
		products = append(products, p)
	}

	if err := cursor.Err(); err != nil {
		log.Printf("Cursor-Fehler: %v", err)
		return nil, err
	}

	log.Printf("%d Produkte gefunden", len(products))
	return products, nil
}

// FindByID sucht ein Produkt anhand seiner ID.
func (r *ProductRepository) FindByID(id string) (*domain.Product, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Printf("Ungueltige ObjectID: %s", id)
		return nil, err
	}

	var product domain.Product
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&product)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Printf("Produkt mit ID %s nicht gefunden", id)
			return nil, err
		}
		log.Printf("Fehler bei Produktsuche: %v", err)
		return nil, err
	}

	log.Printf("Produkt gefunden: %s - %.2f EUR", product.Name, product.Price)
	return &product, nil
}
