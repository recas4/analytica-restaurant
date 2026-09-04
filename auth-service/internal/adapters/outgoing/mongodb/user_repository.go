package mongodb

import (
	"auth-service/internal/application/ports"
	"auth-service/internal/domain"
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Standard-Timeout fuer Datenbank-Operationen.
const defaultTimeout = 5 * time.Second

// MongoUserRepository implementiert UserRepository mit MongoDB als Speicher.
type MongoUserRepository struct {
	collection *mongo.Collection
}

// NewMongoUserRepository erstellt eine neue MongoDB-Repository-Instanz.
func NewMongoUserRepository(db *mongo.Database, collectionName string) ports.UserRepository {
	return &MongoUserRepository{
		collection: db.Collection(collectionName),
	}
}

// Create fuegt ein neues Benutzer-Dokument in die Collection ein.
func (r *MongoUserRepository) Create(user *domain.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	result, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		return err
	}

	// Benutzer-ID mit der generierten MongoDB ObjectID aktualisieren
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		user.ID = oid.Hex()
	}

	return nil
}

// FindByEmail ruft einen Benutzer anhand seiner E-Mail-Adresse ab.
func (r *MongoUserRepository) FindByEmail(email string) (*domain.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	var user domain.User
	filter := bson.M{"email": email}

	err := r.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// FindByID ruft einen Benutzer anhand seiner eindeutigen ID ab.
func (r *MongoUserRepository) FindByID(id string) (*domain.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var user domain.User
	filter := bson.M{"_id": objectID}

	err = r.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// ExistsByEmail prueft ob ein Benutzer mit der angegebenen E-Mail existiert.
func (r *MongoUserRepository) ExistsByEmail(email string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	filter := bson.M{"email": email}
	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
