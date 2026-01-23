package usecase

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/internal/models"
	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/internal/product"
	prodKafka "github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/internal/product/delivery/kafka"
	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/pkg/logger"
	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/pkg/utils"
)

// productUC
type productUC struct {
	productRepo  product.MongoRepository
	redisRepo    product.RedisRepository
	log          logger.Logger
	prodProducer prodKafka.ProductsProducer
}

// NewProductUC constructor
func NewProductUC(
	productRepo product.MongoRepository,
	redisRepo product.RedisRepository,
	log logger.Logger,
	prodProducer prodKafka.ProductsProducer,
) *productUC {
	return &productUC{productRepo: productRepo, redisRepo: redisRepo, log: log, prodProducer: prodProducer}
}

// Create Create new product
func (p *productUC) Create(ctx context.Context, product *models.Product) (*models.Product, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "productUC.Create")
	defer span.Finish()
	return p.productRepo.Create(ctx, product)
}

// Update single product
func (p *productUC) Update(ctx context.Context, product *models.Product) (*models.Product, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "productUC.Update")
	defer span.Finish()

	prod, err := p.productRepo.Update(ctx, product)
	if err != nil {
		return nil, errors.Wrap(err, "Update")
	}

	if err := p.redisRepo.SetProduct(ctx, prod); err != nil {
		p.log.Errorf("redisRepo.SetProduct: %v", err)
	}

	return prod, nil
}

// GetByID Get single product by id
func (p *productUC) GetByID(ctx context.Context, productID primitive.ObjectID) (*models.Product, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "productUC.GetByID")
	defer span.Finish()

	cached, err := p.redisRepo.GetProductByID(ctx, productID)
	if err != nil && err != redis.Nil {
		p.log.Errorf("redisRepo.GetProductByID: %v", err)
	}
	if cached != nil {
		return cached, nil
	}

	prod, err := p.productRepo.GetByID(ctx, productID)
	if err != nil {
		return nil, errors.Wrap(err, "GetByID")
	}

	if err := p.redisRepo.SetProduct(ctx, prod); err != nil {
		p.log.Errorf("redisRepo.SetProduct: %v", err)
	}

	return prod, nil
}

// GetAll Get all products with pagination
func (p *productUC) GetAll(ctx context.Context, pagination *utils.Pagination) (*models.ProductsList, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "productUC.GetAll")
	defer span.Finish()
	return p.productRepo.GetAll(ctx, pagination)
}

// Search Search products
func (p *productUC) Search(ctx context.Context, search string, pagination *utils.Pagination) (*models.ProductsList, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "productUC.Search")
	defer span.Finish()
	return p.productRepo.Search(ctx, search, pagination)
}

// PublishCreate create new product
func (p *productUC) PublishCreate(ctx context.Context, product *models.Product) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "productUC.PublishCreate")
	defer span.Finish()

	prodBytes, err := json.Marshal(&product)
	if err != nil {
		return errors.Wrap(err, "json.Marshal")
	}

	return p.prodProducer.PublishCreate(ctx, kafka.Message{
		Value: prodBytes,
		Time:  time.Now().UTC(),
	})
}

// PublishUpdate update new product
func (p *productUC) PublishUpdate(ctx context.Context, product *models.Product) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "productUC.PublishUpdate")
	defer span.Finish()

	prodBytes, err := json.Marshal(&product)
	if err != nil {
		return errors.Wrap(err, "json.Marshal")
	}

	return p.prodProducer.PublishUpdate(ctx, kafka.Message{
		Value: prodBytes,
		Time:  time.Now().UTC(),
	})
}
