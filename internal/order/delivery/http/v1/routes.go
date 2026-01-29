package v1

// MapRoutes maps order routes
func (h *orderHandlers) MapRoutes() {
	h.group.Post("/", h.CreateOrder())
	h.group.Get("/user/:user_id", h.GetUserOrders()) // Must be before /:order_id
	h.group.Get("/:order_id", h.GetOrder())
	h.group.Put("/:order_id/status", h.UpdateOrderStatus())
	h.group.Put("/:order_id/cancel", h.CancelOrder())
}
