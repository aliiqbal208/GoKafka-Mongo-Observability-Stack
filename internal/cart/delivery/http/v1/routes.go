package v1

// MapRoutes maps cart routes
func (h *cartHandlers) MapRoutes() {
	h.group.Get("/:user_id", h.GetCart())
	h.group.Post("/:user_id/items", h.AddItem())
	h.group.Put("/:user_id/items/:product_id", h.UpdateItem())
	h.group.Delete("/:user_id/items/:product_id", h.RemoveItem())
	h.group.Delete("/:user_id", h.ClearCart())
}
