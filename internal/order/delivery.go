package order

// Handlers defines the order HTTP handlers interface
type Handlers interface {
	CreateOrder() func(c interface{}) error
	GetOrder() func(c interface{}) error
	GetUserOrders() func(c interface{}) error
	UpdateOrderStatus() func(c interface{}) error
	CancelOrder() func(c interface{}) error
}
