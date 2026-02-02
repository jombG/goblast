package shop

import (
	"errors"
	"fmt"
)

// Product represents a product in the shop
type Product struct {
	ID    int
	Name  string
	Price float64
	Stock int
}

// ProductService handles product operations
type ProductService struct {
	products map[int]*Product
	nextID   int
}

// NewProductService creates a new product service
func NewProductService() *ProductService {
	return &ProductService{
		products: make(map[int]*Product),
		nextID:   1,
	}
}

// AddProduct adds a new product to the catalog
func (s *ProductService) AddProduct(name string, price float64, stock int) (*Product, error) {
	if name == "" {
		return nil, errors.New("product name cannot be empty")
	}
	if price < 0 {
		return nil, errors.New("price cannot be negative")
	}
	if stock < 0 {
		return nil, errors.New("stock cannot be negative")
	}

	product := &Product{
		ID:    s.nextID,
		Name:  name,
		Price: price,
		Stock: stock,
	}
	s.products[product.ID] = product
	s.nextID++

	return product, nil
}

// GetProduct retrieves a product by ID
func (s *ProductService) GetProduct(id int) (*Product, error) {
	product, exists := s.products[id]
	if !exists {
		return nil, fmt.Errorf("product %d not found", id)
	}
	return product, nil
}

// UpdateStock updates the stock of a product
func (s *ProductService) UpdateStock(id int, quantity int) error {
	product, err := s.GetProduct(id)
	if err != nil {
		return err
	}

	newStock := product.Stock + quantity
	if newStock < 0 {
		return errors.New("insufficient stock")
	}

	product.Stock = newStock
	return nil
}

// CalculateDiscount calculates discount price
func CalculateDiscount(price float64, percentage int) float64 {
	if percentage < 0 || percentage > 100 {
		return price
	}
	discount := price * float64(percentage) / 100
	return price - discount
}

// FormatPrice formats price as string
func FormatPrice(price float64) string {
	return fmt.Sprintf("$%.2f", price)
}

// internal helper function
func validatePrice(price float64) bool {
	return price >= 0 && price <= 1000000
}
