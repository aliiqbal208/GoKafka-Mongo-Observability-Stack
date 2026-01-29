package repository

import (
	"context"
	"time"

	"github.com/opentracing/opentracing-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/internal/models"
	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/internal/order"
	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/pkg/logger"
)

const (
	ordersCollection = "orders"
)

type orderMongoRepository struct {
	log logger.Logger
	db  *mongo.Database
}

// NewOrderMongoRepository creates a new order MongoDB repository
func NewOrderMongoRepository(log logger.Logger, db *mongo.Database) order.MongoRepository {
	return &orderMongoRepository{log: log, db: db}
}

// Create creates a new order
func (r *orderMongoRepository) Create(ctx context.Context, ord *models.Order) (*models.Order, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "orderMongoRepository.Create")
	defer span.Finish()

	ord.OrderID = primitive.NewObjectID()
	ord.CreatedAt = time.Now().UTC()
	ord.UpdatedAt = time.Now().UTC()
	ord.Status = models.OrderStatusPending

	_, err := r.db.Collection(ordersCollection).InsertOne(ctx, ord)
	if err != nil {
		r.log.Errorf("orderMongoRepository.Create.InsertOne: %v", err)
		return nil, err
	}

	return ord, nil
}

// GetByID retrieves an order by ID
func (r *orderMongoRepository) GetByID(ctx context.Context, orderID string) (*models.Order, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "orderMongoRepository.GetByID")
	defer span.Finish()

	objID, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		return nil, err
	}

	var ord models.Order
	if err := r.db.Collection(ordersCollection).FindOne(ctx, bson.M{"_id": objID}).Decode(&ord); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		r.log.Errorf("orderMongoRepository.GetByID.FindOne: %v", err)
		return nil, err
	}

	return &ord, nil
}

// GetByUserID retrieves all orders for a user with pagination
func (r *orderMongoRepository) GetByUserID(ctx context.Context, userID string, page, size int64) (*models.OrdersList, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "orderMongoRepository.GetByUserID")
	defer span.Finish()

	filter := bson.M{"userId": userID}

	// Count total documents
	totalCount, err := r.db.Collection(ordersCollection).CountDocuments(ctx, filter)
	if err != nil {
		r.log.Errorf("orderMongoRepository.GetByUserID.CountDocuments: %v", err)
		return nil, err
	}

	// Calculate pagination
	skip := (page - 1) * size
	opts := options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: -1}}).
		SetSkip(skip).
		SetLimit(size)

	cursor, err := r.db.Collection(ordersCollection).Find(ctx, filter, opts)
	if err != nil {
		r.log.Errorf("orderMongoRepository.GetByUserID.Find: %v", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var orders []*models.Order
	if err := cursor.All(ctx, &orders); err != nil {
		r.log.Errorf("orderMongoRepository.GetByUserID.All: %v", err)
		return nil, err
	}

	totalPages := totalCount / size
	if totalCount%size != 0 {
		totalPages++
	}

	return &models.OrdersList{
		TotalCount: totalCount,
		TotalPages: totalPages,
		Page:       page,
		Size:       size,
		HasMore:    page < totalPages,
		Orders:     orders,
	}, nil
}

// UpdateStatus updates an order's status
func (r *orderMongoRepository) UpdateStatus(ctx context.Context, orderID string, status models.OrderStatus) (*models.Order, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "orderMongoRepository.UpdateStatus")
	defer span.Finish()

	objID, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		return nil, err
	}

	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now().UTC(),
		},
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updatedOrder models.Order
	err = r.db.Collection(ordersCollection).FindOneAndUpdate(ctx, bson.M{"_id": objID}, update, opts).Decode(&updatedOrder)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		r.log.Errorf("orderMongoRepository.UpdateStatus.FindOneAndUpdate: %v", err)
		return nil, err
	}

	return &updatedOrder, nil
}

// Update updates an order
func (r *orderMongoRepository) Update(ctx context.Context, ord *models.Order) (*models.Order, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "orderMongoRepository.Update")
	defer span.Finish()

	ord.UpdatedAt = time.Now().UTC()

	update := bson.M{
		"$set": bson.M{
			"items":            ord.Items,
			"total_amount":     ord.TotalAmount,
			"status":           ord.Status,
			"shipping_address": ord.ShippingAddress,
			"payment_method":   ord.PaymentMethod,
			"updated_at":       ord.UpdatedAt,
		},
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updatedOrder models.Order
	err := r.db.Collection(ordersCollection).FindOneAndUpdate(ctx, bson.M{"_id": ord.OrderID}, update, opts).Decode(&updatedOrder)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		r.log.Errorf("orderMongoRepository.Update.FindOneAndUpdate: %v", err)
		return nil, err
	}

	return &updatedOrder, nil
}

// Delete deletes an order
func (r *orderMongoRepository) Delete(ctx context.Context, orderID string) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "orderMongoRepository.Delete")
	defer span.Finish()

	objID, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		return err
	}

	_, err = r.db.Collection(ordersCollection).DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		r.log.Errorf("orderMongoRepository.Delete.DeleteOne: %v", err)
		return err
	}

	return nil
}
