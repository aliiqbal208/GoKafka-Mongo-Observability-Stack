package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CartItem represents a single item in the cart
// @Description Cart item with product details
type CartItem struct {
	ProductID   primitive.ObjectID `json:"productId" bson:"productId" validate:"required"`
	ProductName string             `json:"productName" bson:"productName" validate:"required"`
	Price       float64            `json:"price" bson:"price" validate:"required,gt=0"`
	Quantity    int64              `json:"quantity" bson:"quantity" validate:"required,gt=0"`
	ImageURL    string             `json:"imageUrl,omitempty" bson:"imageUrl,omitempty"`
}

// GetSubtotal returns the subtotal for this cart item
func (ci *CartItem) GetSubtotal() float64 {
	return ci.Price * float64(ci.Quantity)
}

// Cart represents a user's shopping cart
// @Description Shopping cart with items
type Cart struct {
	CartID    primitive.ObjectID `json:"cartId" bson:"_id,omitempty"`
	UserID    string             `json:"userId" bson:"userId" validate:"required"`
	Items     []CartItem         `json:"items" bson:"items"`
	Total     float64            `json:"total" bson:"total"`
	ItemCount int                `json:"itemCount" bson:"itemCount"`
	CreatedAt time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time          `json:"updatedAt" bson:"updatedAt"`
}

// CalculateTotal calculates the total price of all items in cart
func (c *Cart) CalculateTotal() float64 {
	var total float64
	for _, item := range c.Items {
		total += item.GetSubtotal()
	}
	c.Total = total
	c.ItemCount = len(c.Items)
	return total
}

// AddItem adds a new item or updates quantity if item exists
func (c *Cart) AddItem(item CartItem) {
	for i, existingItem := range c.Items {
		if existingItem.ProductID == item.ProductID {
			c.Items[i].Quantity += item.Quantity
			c.CalculateTotal()
			return
		}
	}
	c.Items = append(c.Items, item)
	c.CalculateTotal()
}

// UpdateItemQuantity updates the quantity of an item in cart
func (c *Cart) UpdateItemQuantity(productID primitive.ObjectID, quantity int64) bool {
	for i, item := range c.Items {
		if item.ProductID == productID {
			if quantity <= 0 {
				// Remove item if quantity is 0 or less
				c.Items = append(c.Items[:i], c.Items[i+1:]...)
			} else {
				c.Items[i].Quantity = quantity
			}
			c.CalculateTotal()
			return true
		}
	}
	return false
}

// RemoveItem removes an item from the cart
func (c *Cart) RemoveItem(productID primitive.ObjectID) bool {
	for i, item := range c.Items {
		if item.ProductID == productID {
			c.Items = append(c.Items[:i], c.Items[i+1:]...)
			c.CalculateTotal()
			return true
		}
	}
	return false
}

// Clear removes all items from the cart
func (c *Cart) Clear() {
	c.Items = []CartItem{}
	c.Total = 0
	c.ItemCount = 0
}

// IsEmpty checks if cart has no items
func (c *Cart) IsEmpty() bool {
	return len(c.Items) == 0
}

// ToOrderItems converts cart items to order items
func (c *Cart) ToOrderItems() []OrderItem {
	orderItems := make([]OrderItem, len(c.Items))
	for i, item := range c.Items {
		orderItems[i] = OrderItem{
			ProductID:   item.ProductID,
			ProductName: item.ProductName,
			Price:       item.Price,
			Quantity:    item.Quantity,
			Subtotal:    item.GetSubtotal(),
		}
	}
	return orderItems
}

// AddItemRequest represents the request to add item to cart
// @Description Request body for adding item to cart
type AddItemRequest struct {
	ProductID string `json:"productId" validate:"required" example:"507f1f77bcf86cd799439011"`
	Quantity  int64  `json:"quantity" validate:"required,gt=0" example:"2"`
}

// UpdateItemRequest represents the request to update item quantity
// @Description Request body for updating item quantity
type UpdateItemRequest struct {
	Quantity int64 `json:"quantity" validate:"required,gt=0" example:"3"`
}
