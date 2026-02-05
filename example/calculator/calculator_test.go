package calculator

import (
	"testing"

	"jombG/goblast/example/shop"
)

func TestAdd(t *testing.T) {
	result := Add(2, 3)
	expected := 5
	if result != expected {
		t.Errorf("Add(2, 3) = %d; want %d", result, expected)
	}
}

func TestSubtract(t *testing.T) {
	result := Subtract(5, 3)
	expected := 2
	if result != expected {
		t.Errorf("Subtract(5, 3) = %d; want %d", result, expected)
	}
}

func TestAddPriceProdcut(t *testing.T) {
	pr1 := shop.Product{ID: 1, Name: "Product 1", Price: 100.0, Stock: 10}
	pr2 := shop.Product{ID: 2, Name: "Product 2", Price: 50.0, Stock: 5}

	result := AddPriceProdcut(pr1, pr2)

	expected := 150.0
	if result.Price != expected {
		t.Errorf("AddPriceProdcut() price = %.2f; want %.2f", result.Price, expected)
	}

	if result.ID != pr1.ID {
		t.Errorf("AddPriceProdcut() ID = %d; want %d", result.ID, pr1.ID)
	}
	if result.Name != pr1.Name {
		t.Errorf("AddPriceProdcut() Name = %s; want %s", result.Name, pr1.Name)
	}
}
