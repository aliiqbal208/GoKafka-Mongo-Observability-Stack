package kafka

// Kafka topics for order events
const (
	OrderCreatedTopic   = "order-created"
	OrderUpdatedTopic   = "order-updated"
	OrderCancelledTopic = "order-cancelled"

	OrderConsumerGroup = "order-consumer-group"

	// Pool sizes
	CreateOrderWorkers = 4
	UpdateOrderWorkers = 2
)
