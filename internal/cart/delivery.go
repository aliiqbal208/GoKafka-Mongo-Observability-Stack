package cart

import "github.com/gofiber/fiber/v2"

// Handlers defines the cart HTTP handlers interface
type Handlers interface {
	// GetCart gets a user's cart
	GetCart() fiber.Handler

	// AddItem adds an item to the cart
	AddItem() fiber.Handler

	// UpdateItem updates an item quantity
	UpdateItem() fiber.Handler

	// RemoveItem removes an item from cart
	RemoveItem() fiber.Handler

	// ClearCart clears all items from cart
	ClearCart() fiber.Handler

	// MapRoutes maps the cart routes
	MapRoutes()
}
