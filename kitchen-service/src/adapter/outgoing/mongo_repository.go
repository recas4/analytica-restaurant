package outgoing

import (
	"context"
	"fmt"
	"kitchen-service/src/application/ports"
	"kitchen-service/src/domain"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoRepository struct {
	collection *mongo.Collection
}

func NewMongoRepository(database *mongo.Database) ports.KitchenRepository {
	return &MongoRepository{
		collection: database.Collection("kitchen_orders"),
	}
}

func (kr *MongoRepository) SaveOrder(ctx context.Context, order *domain.KitchenOrder) error {
	if order.ID.IsZero() {
		order.ID = primitive.NewObjectID()
	}
	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()

	_, err := kr.collection.InsertOne(ctx, order)
	if err != nil {
		return fmt.Errorf("failed to save kitchen order: %w", err)
	}

	return nil
}

func (kr *MongoRepository) UpdateOrder(ctx context.Context, order *domain.KitchenOrder) error {
	order.UpdatedAt = time.Now()

	filter := bson.M{"order_id": order.OrderID}
	update := bson.M{"$set": order}

	result, err := kr.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update kitchen order: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("kitchen order with ID %s not found", order.OrderID)
	}

	return nil
}

func (kr *MongoRepository) GetOrderByID(ctx context.Context, orderID string) (*domain.KitchenOrder, error) {
	var order domain.KitchenOrder
	filter := bson.M{"order_id": orderID}

	err := kr.collection.FindOne(ctx, filter).Decode(&order)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("kitchen order with ID %s not found", orderID)
		}
		return nil, fmt.Errorf("failed to get kitchen order: %w", err)
	}

	return &order, nil
}

func (kr *MongoRepository) GetOrdersByStatus(ctx context.Context, status domain.OrderStatus) ([]*domain.KitchenOrder, error) {
	filter := bson.M{"status": status}
	opts := options.Find().SetSort(bson.D{{"created_at", 1}})

	cursor, err := kr.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find orders by status: %w", err)
	}
	defer cursor.Close(ctx)

	var orders []*domain.KitchenOrder
	for cursor.Next(ctx) {
		var order domain.KitchenOrder
		if err := cursor.Decode(&order); err != nil {
			return nil, fmt.Errorf("failed to decode kitchen order: %w", err)
		}
		orders = append(orders, &order)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return orders, nil
}

func (kr *MongoRepository) GetAllOrders(ctx context.Context) ([]*domain.KitchenOrder, error) {
	opts := options.Find().SetSort(bson.D{{"created_at", -1}})

	cursor, err := kr.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find all orders: %w", err)
	}
	defer cursor.Close(ctx)

	var orders []*domain.KitchenOrder
	for cursor.Next(ctx) {
		var order domain.KitchenOrder
		if err := cursor.Decode(&order); err != nil {
			return nil, fmt.Errorf("failed to decode kitchen order: %w", err)
		}
		orders = append(orders, &order)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return orders, nil
}

func (kr *MongoRepository) DeleteOrder(ctx context.Context, orderID string) error {
	filter := bson.M{"order_id": orderID}

	result, err := kr.collection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete kitchen order: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("kitchen order with ID %s not found", orderID)
	}

	return nil
}

func (kr *MongoRepository) GetStats(ctx context.Context) (*domain.KitchenStats, error) {
	pipeline := []bson.M{
		{
			"$group": bson.M{
				"_id": "$status",
				"count": bson.M{"$sum": 1},
				"avgTime": bson.M{"$avg": "$estimated_time"},
			},
		},
	}

	cursor, err := kr.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate stats: %w", err)
	}
	defer cursor.Close(ctx)

	stats := &domain.KitchenStats{}
	statusCounts := make(map[domain.OrderStatus]int)
	totalTime := 0.0
	totalOrders := 0

	for cursor.Next(ctx) {
		var result struct {
			ID      domain.OrderStatus `bson:"_id"`
			Count   int                `bson:"count"`
			AvgTime float64            `bson:"avgTime"`
		}

		if err := cursor.Decode(&result); err != nil {
			return nil, fmt.Errorf("failed to decode stats result: %w", err)
		}

		statusCounts[result.ID] = result.Count
		totalOrders += result.Count
		totalTime += result.AvgTime * float64(result.Count)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	stats.TotalOrders = totalOrders
	stats.OrdersReceived = statusCounts[domain.StatusReceived]
	stats.OrdersPreparing = statusCounts[domain.StatusPreparing]
	stats.OrdersReady = statusCounts[domain.StatusReady]
	stats.OrdersPickedUpByDriver = statusCounts[domain.StatusPickedUpByDriver]

	if totalOrders > 0 {
		stats.AverageWaitTime = int(totalTime / float64(totalOrders))
	}

	return stats, nil
}
