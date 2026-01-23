package product

import "github.com/gofiber/fiber/v2"

// HttpDelivery http delivery
type HttpDelivery interface {
	CreateProduct() fiber.Handler
	UpdateProduct() fiber.Handler
	GetByIDProduct() fiber.Handler
	GetAllProducts() fiber.Handler
	SearchProduct() fiber.Handler
}
