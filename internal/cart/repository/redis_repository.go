package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"

	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/internal/cart"
	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/internal/models"
	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/pkg/logger"
)

const (
	cartCachePrefix = "cart:"
	cartCacheTTL    = time.Hour * 24 // Cart cache expires in 24 hours
)

// cartRedisRepository implements cart.RedisRepository
type cartRedisRepository struct {
	log         logger.Logger
	redisClient *redis.Client
}

// NewCartRedisRepository creates a new cart Redis repository
func NewCartRedisRepository(log logger.Logger, redisClient *redis.Client) cart.RedisRepository {
	return &cartRedisRepository{log: log, redisClient: redisClient}
}

// getCacheKey returns the cache key for a user's cart
func (r *cartRedisRepository) getCacheKey(userID string) string {
	return cartCachePrefix + userID
}

// GetByUserID retrieves a cart from cache
func (r *cartRedisRepository) GetByUserID(ctx context.Context, userID string) (*models.Cart, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "cartRedisRepository.GetByUserID")
	defer span.Finish()

	cacheKey := r.getCacheKey(userID)
	data, err := r.redisClient.Get(ctx, cacheKey).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil // Cache miss
		}
		r.log.Errorf("Redis Get error: %v", err)
		return nil, errors.Wrap(err, "redisClient.Get")
	}

	var cartDoc models.Cart
	if err := json.Unmarshal(data, &cartDoc); err != nil {
		r.log.Errorf("JSON Unmarshal error: %v", err)
		return nil, errors.Wrap(err, "json.Unmarshal")
	}

	return &cartDoc, nil
}

// SetCart caches a cart
func (r *cartRedisRepository) SetCart(ctx context.Context, cart *models.Cart) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "cartRedisRepository.SetCart")
	defer span.Finish()

	cacheKey := r.getCacheKey(cart.UserID)
	data, err := json.Marshal(cart)
	if err != nil {
		r.log.Errorf("JSON Marshal error: %v", err)
		return errors.Wrap(err, "json.Marshal")
	}

	err = r.redisClient.Set(ctx, cacheKey, data, cartCacheTTL).Err()
	if err != nil {
		r.log.Errorf("Redis Set error: %v", err)
		return errors.Wrap(err, "redisClient.Set")
	}

	return nil
}

// DeleteCart removes a cart from cache
func (r *cartRedisRepository) DeleteCart(ctx context.Context, userID string) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "cartRedisRepository.DeleteCart")
	defer span.Finish()

	cacheKey := r.getCacheKey(userID)
	err := r.redisClient.Del(ctx, cacheKey).Err()
	if err != nil {
		r.log.Errorf("Redis Del error: %v", err)
		return errors.Wrap(err, "redisClient.Del")
	}

	return nil
}
