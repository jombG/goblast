package shop

import (
	"errors"
	"fmt"
)

// CartItem represents an item in the shopping cart
type CartItem struct {
	Product  *Product
	Quantity int
}

// Cart represents a shopping cart
type Cart struct {
	Items []*CartItem
}

// NewCart creates a new empty cart
func NewCart() *Cart {
	return &Cart{
		Items: make([]*CartItem, 0),
	}
}

// AddItem adds a product to the cart
func (c *Cart) AddItem(product *Product, quantity int) error {
	if product == nil {
		return errors.New("product cannot be nil")
	}
	if quantity <= 0 {
		return errors.New("quantity must be positive")
	}
	if quantity > product.Stock {
		return fmt.Errorf("insufficient stock: requested %d, available %d", quantity, product.Stock)
	}

	// Check if product already in cart
	for _, item := range c.Items {
		if item.Product.ID == product.ID {
			item.Quantity += quantity
			return nil
		}
	}

	// Add new item
	c.Items = append(c.Items, &CartItem{
		Product:  product,
		Quantity: quantity,
	})

	return nil
}

// RemoveItem removes a product from the cart
func (c *Cart) RemoveItem(productID int) error {
	for i, item := range c.Items {
		if item.Product.ID == productID {
			c.Items = append(c.Items[:i], c.Items[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("product %d not in cart", productID)
}

// GetTotal calculates the total price of items in the cart
func (c *Cart) GetTotal() float64 {
	total := 0.0
	for _, item := range c.Items {
		total += item.Product.Price * float64(item.Quantity)
	}
	return total
}

// GetTotalWithDiscount calculates total with discount
func (c *Cart) GetTotalWithDiscount(discountPercent int) float64 {
	total := c.GetTotal()
	return CalculateDiscount(total, discountPercent)
}

// ItemCount returns the number of different products in cart
func (c *Cart) ItemCount() int {
	return len(c.Items)
}

// TotalQuantity returns the total quantity of all items
func (c *Cart) TotalQuantity() int {
	total := 0
	for _, item := range c.Items {
		total += item.Quantity
	}
	return total
}

// Clear removes all items from the cart
func (c *Cart) Clear() {
	c.Items = make([]*CartItem, 0)
}
