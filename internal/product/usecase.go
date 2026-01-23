package product

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/internal/models"
	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/pkg/utils"
)

// UseCase Product
type UseCase interface {
	Create(ctx context.Context, product *models.Product) (*models.Product, error)
	Update(ctx context.Context, product *models.Product) (*models.Product, error)
	GetByID(ctx context.Context, productID primitive.ObjectID) (*models.Product, error)
	GetAll(ctx context.Context, pagination *utils.Pagination) (*models.ProductsList, error)
	Search(ctx context.Context, search string, pagination *utils.Pagination) (*models.ProductsList, error)
	PublishCreate(ctx context.Context, product *models.Product) error
	PublishUpdate(ctx context.Context, product *models.Product) error
}
