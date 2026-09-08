package domain

import (
	"testing"

	"github.com/google/uuid"
)

func TestCart_AddItem_SameProductMergesQuantity(t *testing.T) {
	cart := NewCart(uuid.New())
	productID := uuid.New()
	_ = cart.AddItem(productID, "Kemeja", 50_000, 2)
	_ = cart.AddItem(productID, "Kemeja", 50_000, 3)
	if len(cart.Items()) != 1 {
		t.Fatalf("expected 1 distinct item, got: %d", len(cart.Items()))
	}
	if cart.Items()[0].Quantity() != 5 {
		t.Errorf("expected quantity 5, got: %d", cart.Items()[0].Quantity())
	}
}

func TestCart_RemoveItem_NotFound(t *testing.T) {
	cart := NewCart(uuid.New())
	if err := cart.RemoveItem(uuid.New()); err != ErrCartItemNotFound {
		t.Errorf("expected ErrCartItemNotFound, got: %v", err)
	}
}

func TestCart_Clear(t *testing.T) {
	cart := NewCart(uuid.New())
	_ = cart.AddItem(uuid.New(), "A", 10_000, 1)
	cart.Clear()
	if len(cart.Items()) != 0 || cart.Total() != 0 {
		t.Errorf("expected empty cart after Clear")
	}
}
