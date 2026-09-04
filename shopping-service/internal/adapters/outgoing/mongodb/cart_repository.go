package mongodb

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"shopping-service/internal/application/ports"
	"shopping-service/internal/domain/models"
)

// CartRepository implementiert das CartRepository-Interface mit MongoDB.
type CartRepository struct {
	collection *mongo.Collection
}

// NewCartRepository erstellt eine neue CartRepository-Instanz.
func NewCartRepository(db *mongo.Database) ports.CartRepository {
	return &CartRepository{
		collection: db.Collection("carts"),
	}
}

// AddItem fuegt einen Artikel zum Warenkorb hinzu.
func (r *CartRepository) AddItem(userID, productID string, qty int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"user_id": userID}
	update := bson.M{"$push": bson.M{"items": bson.M{"product_id": productID, "qty": qty}}}
	upsert := true

	_, err := r.collection.UpdateOne(ctx, filter, update, &options.UpdateOptions{Upsert: &upsert})
	return err
}

// GetCart liefert den Warenkorb eines Benutzers.
func (r *CartRepository) GetCart(userID string) (domain.Cart, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var doc struct {
		UserID string `bson:"user_id"`
		Items  []struct {
			ProductID string `bson:"product_id"`
			Qty       int    `bson:"qty"`
		} `bson:"items"`
	}

	err := r.collection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return domain.Cart{UserID: userID, Items: []domain.CartItem{}}, nil
		}
		return domain.Cart{}, err
	}

	cart := domain.Cart{UserID: doc.UserID}
	for _, item := range doc.Items {
		cart.Items = append(cart.Items, domain.CartItem{
			ProductID: item.ProductID,
			Qty:       item.Qty,
		})
	}

	return cart, nil
}

// ClearCart leert den Warenkorb eines Benutzers.
func (r *CartRepository) ClearCart(userID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := r.collection.DeleteOne(ctx, bson.M{"user_id": userID})
	return err
}

// UpdateItem aktualisiert die Menge eines Artikels.
func (r *CartRepository) UpdateItem(userID, productID string, qty int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{
		"user_id":          userID,
		"items.product_id": productID,
	}
	update := bson.M{
		"$set": bson.M{"items.$.qty": qty},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	// Falls Artikel nicht gefunden, fuege ihn hinzu
	if result.ModifiedCount == 0 {
		return r.AddItem(userID, productID, qty)
	}

	return nil
}

// RemoveItem entfernt einen Artikel aus dem Warenkorb.
func (r *CartRepository) RemoveItem(userID, productID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"user_id": userID}
	update := bson.M{
		"$pull": bson.M{"items": bson.M{"product_id": productID}},
	}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}
