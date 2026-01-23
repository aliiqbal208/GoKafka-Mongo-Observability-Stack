package v1

import (
	"github.com/gofiber/fiber/v2"
)

// MapRoutes products routes
func (p *productHandlers) MapRoutes() {

	p.group.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("Hello from test endpoint!")
	})
	p.group.Post("", p.CreateProduct())
	p.group.Put("/:product_id", p.UpdateProduct())
	p.group.Get("/search", p.SearchProduct())
	p.group.Get("/:product_id", p.GetByIDProduct())
	p.group.Get("", p.GetAllProducts())
}
