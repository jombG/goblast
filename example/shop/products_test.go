package shop

import (
	"testing"
)

func TestNewProductService(t *testing.T) {
	service := NewProductService()
	if service == nil {
		t.Fatal("expected non-nil service")
	}
	if service.nextID != 1 {
		t.Errorf("expected nextID to be 1, got %d", service.nextID)
	}
}

func TestAddProduct(t *testing.T) {
	service := NewProductService()

	tests := []struct {
		name      string
		prodName  string
		price     float64
		stock     int
		wantError bool
	}{
		{"valid product", "Laptop", 999.99, 10, false},
		{"empty name", "", 100, 5, true},
		{"negative price", "Phone", -50, 5, true},
		{"negative stock", "Tablet", 200, -1, true},
		{"zero price", "Free Item", 0, 100, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			product, err := service.AddProduct(tt.prodName, tt.price, tt.stock)
			if tt.wantError {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if product.Name != tt.prodName {
					t.Errorf("expected name %s, got %s", tt.prodName, product.Name)
				}
			}
		})
	}
}

func TestGetProduct(t *testing.T) {
	service := NewProductService()
	product, _ := service.AddProduct("Keyboard", 79.99, 20)

	// Test existing product
	retrieved, err := service.GetProduct(product.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if retrieved.Name != "Keyboard" {
		t.Errorf("expected Keyboard, got %s", retrieved.Name)
	}

	// Test non-existing product
	_, err = service.GetProduct(999)
	if err == nil {
		t.Error("expected error for non-existing product")
	}
}

func TestUpdateStock(t *testing.T) {
	service := NewProductService()
	product, _ := service.AddProduct("Mouse", 29.99, 50)

	// Increase stock
	err := service.UpdateStock(product.ID, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	updated, _ := service.GetProduct(product.ID)
	if updated.Stock != 60 {
		t.Errorf("expected stock 60, got %d", updated.Stock)
	}

	// Decrease stock
	err = service.UpdateStock(product.ID, -30)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	updated, _ = service.GetProduct(product.ID)
	if updated.Stock != 30 {
		t.Errorf("expected stock 30, got %d", updated.Stock)
	}

	// Try to decrease below zero
	err = service.UpdateStock(product.ID, -100)
	if err == nil {
		t.Error("expected error for insufficient stock")
	}
}

func TestCalculateDiscount(t *testing.T) {
	tests := []struct {
		price      float64
		percentage int
		expected   float64
	}{
		{100.0, 10, 90.0},
		{100.0, 50, 50.0},
		{100.0, 0, 100.0},
		{100.0, 100, 0.0},
		{100.0, -10, 100.0}, // invalid
		{100.0, 150, 100.0}, // invalid
	}

	for _, tt := range tests {
		result := CalculateDiscount(tt.price, tt.percentage)
		if result != tt.expected {
			t.Errorf("CalculateDiscount(%f, %d) = %f; want %f",
				tt.price, tt.percentage, result, tt.expected)
		}
	}
}

func TestFormatPrice(t *testing.T) {
	tests := []struct {
		price    float64
		expected string
	}{
		{99.99, "$99.99"},
		{0.0, "$0.00"},
		{1234.56, "$1234.56"},
	}

	for _, tt := range tests {
		result := FormatPrice(tt.price)
		if result != tt.expected {
			t.Errorf("FormatPrice(%f) = %s; want %s",
				tt.price, result, tt.expected)
		}
	}
}
