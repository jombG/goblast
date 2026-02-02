package shop

import (
	"testing"
)

func TestNewCart(t *testing.T) {
	cart := NewCart()
	if cart == nil {
		t.Fatal("expected non-nil cart")
	}
	if cart.ItemCount() != 0 {
		t.Errorf("expected empty cart, got %d items", cart.ItemCount())
	}
}

func TestAddItem(t *testing.T) {
	cart := NewCart()
	product := &Product{ID: 1, Name: "Test", Price: 10.0, Stock: 5}

	// Add valid item
	err := cart.AddItem(product, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cart.ItemCount() != 1 {
		t.Errorf("expected 1 item, got %d", cart.ItemCount())
	}

	// Add same product again (should increase quantity)
	err = cart.AddItem(product, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cart.ItemCount() != 1 {
		t.Errorf("expected still 1 item type, got %d", cart.ItemCount())
	}
	if cart.Items[0].Quantity != 3 {
		t.Errorf("expected quantity 3, got %d", cart.Items[0].Quantity)
	}

	// Try to add nil product
	err = cart.AddItem(nil, 1)
	if err == nil {
		t.Error("expected error for nil product")
	}

	// Try to add with invalid quantity
	err = cart.AddItem(product, 0)
	if err == nil {
		t.Error("expected error for zero quantity")
	}

	// Try to add more than stock
	err = cart.AddItem(product, 10)
	if err == nil {
		t.Error("expected error for insufficient stock")
	}
}

func TestRemoveItem(t *testing.T) {
	cart := NewCart()
	product := &Product{ID: 1, Name: "Test", Price: 10.0, Stock: 5}
	cart.AddItem(product, 2)

	// Remove existing item
	err := cart.RemoveItem(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cart.ItemCount() != 0 {
		t.Errorf("expected empty cart, got %d items", cart.ItemCount())
	}

	// Try to remove non-existing item
	err = cart.RemoveItem(999)
	if err == nil {
		t.Error("expected error for non-existing product")
	}
}

func TestGetTotal(t *testing.T) {
	cart := NewCart()
	product1 := &Product{ID: 1, Name: "Item1", Price: 10.0, Stock: 10}
	product2 := &Product{ID: 2, Name: "Item2", Price: 25.0, Stock: 10}

	cart.AddItem(product1, 2)  // 20.0
	cart.AddItem(product2, 3)  // 75.0
	// Total: 95.0

	total := cart.GetTotal()
	expected := 95.0
	if total != expected {
		t.Errorf("expected total %f, got %f", expected, total)
	}
}

func TestGetTotalWithDiscount(t *testing.T) {
	cart := NewCart()
	product := &Product{ID: 1, Name: "Item", Price: 100.0, Stock: 10}
	cart.AddItem(product, 1) // 100.0

	// 10% discount
	total := cart.GetTotalWithDiscount(10)
	expected := 90.0
	if total != expected {
		t.Errorf("expected total %f, got %f", expected, total)
	}

	// 50% discount
	total = cart.GetTotalWithDiscount(50)
	expected = 50.0
	if total != expected {
		t.Errorf("expected total %f, got %f", expected, total)
	}
}

func TestItemCount(t *testing.T) {
	cart := NewCart()
	product1 := &Product{ID: 1, Name: "Item1", Price: 10.0, Stock: 10}
	product2 := &Product{ID: 2, Name: "Item2", Price: 25.0, Stock: 10}

	if cart.ItemCount() != 0 {
		t.Error("expected 0 items in new cart")
	}

	cart.AddItem(product1, 5)
	if cart.ItemCount() != 1 {
		t.Errorf("expected 1 item type, got %d", cart.ItemCount())
	}

	cart.AddItem(product2, 3)
	if cart.ItemCount() != 2 {
		t.Errorf("expected 2 item types, got %d", cart.ItemCount())
	}
}

func TestTotalQuantity(t *testing.T) {
	cart := NewCart()
	product1 := &Product{ID: 1, Name: "Item1", Price: 10.0, Stock: 10}
	product2 := &Product{ID: 2, Name: "Item2", Price: 25.0, Stock: 10}

	cart.AddItem(product1, 5)
	cart.AddItem(product2, 3)

	total := cart.TotalQuantity()
	expected := 8
	if total != expected {
		t.Errorf("expected total quantity %d, got %d", expected, total)
	}
}

func TestClear(t *testing.T) {
	cart := NewCart()
	product := &Product{ID: 1, Name: "Item", Price: 10.0, Stock: 10}
	cart.AddItem(product, 2)

	cart.Clear()

	if cart.ItemCount() != 0 {
		t.Errorf("expected empty cart after clear, got %d items", cart.ItemCount())
	}
	if cart.GetTotal() != 0 {
		t.Errorf("expected zero total after clear, got %f", cart.GetTotal())
	}
}
