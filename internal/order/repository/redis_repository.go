package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/opentracing/opentracing-go"

	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/internal/models"
	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/internal/order"
	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/pkg/logger"
)

const (
	orderPrefix   = "order:"
	orderCacheTTL = 30 * time.Minute
)

type orderRedisRepository struct {
	log         logger.Logger
	redisClient *redis.Client
}

// NewOrderRedisRepository creates a new order Redis repository
func NewOrderRedisRepository(log logger.Logger, redisClient *redis.Client) order.RedisRepository {
	return &orderRedisRepository{log: log, redisClient: redisClient}
}

func (r *orderRedisRepository) getOrderKey(orderID string) string {
	return fmt.Sprintf("%s%s", orderPrefix, orderID)
}

// GetByID retrieves an order from cache
func (r *orderRedisRepository) GetByID(ctx context.Context, orderID string) (*models.Order, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "orderRedisRepository.GetByID")
	defer span.Finish()

	result, err := r.redisClient.Get(ctx, r.getOrderKey(orderID)).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		r.log.Errorf("orderRedisRepository.GetByID.Get: %v", err)
		return nil, err
	}

	var ord models.Order
	if err := json.Unmarshal([]byte(result), &ord); err != nil {
		r.log.Errorf("orderRedisRepository.GetByID.Unmarshal: %v", err)
		return nil, err
	}

	return &ord, nil
}

// SetOrder caches an order
func (r *orderRedisRepository) SetOrder(ctx context.Context, ord *models.Order) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "orderRedisRepository.SetOrder")
	defer span.Finish()

	data, err := json.Marshal(ord)
	if err != nil {
		r.log.Errorf("orderRedisRepository.SetOrder.Marshal: %v", err)
		return err
	}

	if err := r.redisClient.Set(ctx, r.getOrderKey(ord.OrderID.Hex()), data, orderCacheTTL).Err(); err != nil {
		r.log.Errorf("orderRedisRepository.SetOrder.Set: %v", err)
		return err
	}

	return nil
}

// DeleteOrder removes an order from cache
func (r *orderRedisRepository) DeleteOrder(ctx context.Context, orderID string) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "orderRedisRepository.DeleteOrder")
	defer span.Finish()

	if err := r.redisClient.Del(ctx, r.getOrderKey(orderID)).Err(); err != nil {
		r.log.Errorf("orderRedisRepository.DeleteOrder.Del: %v", err)
		return err
	}

	return nil
}
