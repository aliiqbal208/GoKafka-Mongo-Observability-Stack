package cart

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/internal/models"
)

// UseCase defines the cart business logic interface
type UseCase interface {
	// GetCart retrieves a user's cart
	GetCart(ctx context.Context, userID string) (*models.Cart, error)

	// AddItem adds an item to the cart
	AddItem(ctx context.Context, userID string, productID primitive.ObjectID, quantity int64) (*models.Cart, error)

	// UpdateItemQuantity updates the quantity of an item
	UpdateItemQuantity(ctx context.Context, userID string, productID primitive.ObjectID, quantity int64) (*models.Cart, error)

	// RemoveItem removes an item from the cart
	RemoveItem(ctx context.Context, userID string, productID primitive.ObjectID) (*models.Cart, error)

	// ClearCart removes all items from the cart
	ClearCart(ctx context.Context, userID string) error
}
