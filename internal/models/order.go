package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// OrderStatus represents the status of an order
type OrderStatus string

const (
	OrderStatusPending    OrderStatus = "pending"
	OrderStatusConfirmed  OrderStatus = "confirmed"
	OrderStatusProcessing OrderStatus = "processing"
	OrderStatusShipped    OrderStatus = "shipped"
	OrderStatusDelivered  OrderStatus = "delivered"
	OrderStatusCancelled  OrderStatus = "cancelled"
	OrderStatusRefunded   OrderStatus = "refunded"
)

// IsValid checks if the order status is valid
func (s OrderStatus) IsValid() bool {
	switch s {
	case OrderStatusPending, OrderStatusConfirmed, OrderStatusProcessing,
		OrderStatusShipped, OrderStatusDelivered, OrderStatusCancelled, OrderStatusRefunded:
		return true
	}
	return false
}

// CanTransitionTo checks if status can transition to a new status
func (s OrderStatus) CanTransitionTo(newStatus OrderStatus) bool {
	transitions := map[OrderStatus][]OrderStatus{
		OrderStatusPending:    {OrderStatusConfirmed, OrderStatusCancelled},
		OrderStatusConfirmed:  {OrderStatusProcessing, OrderStatusCancelled},
		OrderStatusProcessing: {OrderStatusShipped, OrderStatusCancelled},
		OrderStatusShipped:    {OrderStatusDelivered},
		OrderStatusDelivered:  {OrderStatusRefunded},
		OrderStatusCancelled:  {},
		OrderStatusRefunded:   {},
	}

	allowedTransitions, exists := transitions[s]
	if !exists {
		return false
	}

	for _, allowed := range allowedTransitions {
		if allowed == newStatus {
			return true
		}
	}
	return false
}

// OrderItem represents a single item in an order
// @Description Order item with product details
type OrderItem struct {
	ProductID   primitive.ObjectID `json:"productId" bson:"productId"`
	ProductName string             `json:"productName" bson:"productName"`
	Price       float64            `json:"price" bson:"price"`
	Quantity    int64              `json:"quantity" bson:"quantity"`
	Subtotal    float64            `json:"subtotal" bson:"subtotal"`
}

// Order represents a customer order
// @Description Customer order with items and status
type Order struct {
	OrderID         primitive.ObjectID `json:"orderId" bson:"_id,omitempty"`
	UserID          string             `json:"userId" bson:"userId" validate:"required"`
	Items           []OrderItem        `json:"items" bson:"items" validate:"required,min=1"`
	TotalAmount     float64            `json:"totalAmount" bson:"totalAmount"`
	Status          OrderStatus        `json:"status" bson:"status"`
	ShippingAddress ShippingAddress    `json:"shippingAddress" bson:"shippingAddress" validate:"required"`
	PaymentMethod   string             `json:"paymentMethod" bson:"paymentMethod" validate:"required"`
	PaymentStatus   string             `json:"paymentStatus" bson:"paymentStatus"`
	Notes           string             `json:"notes,omitempty" bson:"notes,omitempty"`
	CreatedAt       time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt       time.Time          `json:"updatedAt" bson:"updatedAt"`
	ConfirmedAt     *time.Time         `json:"confirmedAt,omitempty" bson:"confirmedAt,omitempty"`
	ShippedAt       *time.Time         `json:"shippedAt,omitempty" bson:"shippedAt,omitempty"`
	DeliveredAt     *time.Time         `json:"deliveredAt,omitempty" bson:"deliveredAt,omitempty"`
}

// ShippingAddress represents the delivery address
// @Description Shipping address details
type ShippingAddress struct {
	FullName    string `json:"fullName" bson:"fullName" validate:"required" example:"John Doe"`
	Phone       string `json:"phone" bson:"phone" validate:"required" example:"+1234567890"`
	AddressLine string `json:"addressLine" bson:"addressLine" validate:"required" example:"123 Main St"`
	City        string `json:"city" bson:"city" validate:"required" example:"New York"`
	State       string `json:"state" bson:"state" validate:"required" example:"NY"`
	PostalCode  string `json:"postalCode" bson:"postalCode" validate:"required" example:"10001"`
	Country     string `json:"country" bson:"country" validate:"required" example:"USA"`
}

// CalculateTotal calculates the total amount of the order
func (o *Order) CalculateTotal() float64 {
	var total float64
	for i, item := range o.Items {
		o.Items[i].Subtotal = item.Price * float64(item.Quantity)
		total += o.Items[i].Subtotal
	}
	o.TotalAmount = total
	return total
}

// UpdateStatus updates the order status with timestamp
func (o *Order) UpdateStatus(newStatus OrderStatus) bool {
	if !o.Status.CanTransitionTo(newStatus) {
		return false
	}

	now := time.Now()
	o.Status = newStatus
	o.UpdatedAt = now

	switch newStatus {
	case OrderStatusConfirmed:
		o.ConfirmedAt = &now
	case OrderStatusShipped:
		o.ShippedAt = &now
	case OrderStatusDelivered:
		o.DeliveredAt = &now
	}

	return true
}

// OrdersList represents a paginated list of orders
// @Description Paginated list of orders
type OrdersList struct {
	TotalCount int64    `json:"totalCount"`
	TotalPages int64    `json:"totalPages"`
	Page       int64    `json:"page"`
	Size       int64    `json:"size"`
	HasMore    bool     `json:"hasMore"`
	Orders     []*Order `json:"orders"`
}

// CreateOrderRequest represents the request to create an order
// @Description Request body for creating an order
type CreateOrderRequest struct {
	UserID          string          `json:"userId" validate:"required" example:"user123"`
	ShippingAddress ShippingAddress `json:"shippingAddress" validate:"required"`
	PaymentMethod   string          `json:"paymentMethod" validate:"required" example:"credit_card"`
	Notes           string          `json:"notes,omitempty" example:"Please deliver after 5 PM"`
}

// UpdateOrderStatusRequest represents the request to update order status
// @Description Request body for updating order status
type UpdateOrderStatusRequest struct {
	Status OrderStatus `json:"status" validate:"required" example:"confirmed"`
}

// OrderEvent represents an order event for Kafka
type OrderEvent struct {
	EventType string    `json:"eventType"`
	OrderID   string    `json:"orderId"`
	UserID    string    `json:"userId"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Data      *Order    `json:"data,omitempty"`
}
