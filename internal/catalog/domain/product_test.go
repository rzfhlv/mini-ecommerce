package domain

import (
	"testing"

	"github.com/google/uuid"
)

func TestProduct_DecreaseStock_Success(t *testing.T) {
	price, _ := NewPrice(50_000)
	product, _ := NewProduct(uuid.New(), "Kemeja Flannel", "desc", price, 10, "M", "")
	if err := product.DecreaseStock(3); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if product.Stock() != 7 {
		t.Errorf("expected stock 7, got: %d", product.Stock())
	}
}

func TestProduct_DecreaseStock_RejectsInsufficientStock(t *testing.T) {
	price, _ := NewPrice(50_000)
	product, _ := NewProduct(uuid.New(), "Celana Chino", "desc", price, 2, "L", "")
	err := product.DecreaseStock(5)
	if err != ErrInsufficientStock {
		t.Errorf("expected ErrInsufficientStock, got: %v", err)
	}
	if product.Stock() != 2 {
		t.Errorf("expected stock unchanged, got: %d", product.Stock())
	}
}

func TestNewProduct_RejectsMissingCategory(t *testing.T) {
	price, _ := NewPrice(50_000)
	_, err := NewProduct(uuid.Nil, "Kemeja", "desc", price, 10, "M", "")
	if err != ErrCategoryRequired {
		t.Errorf("expected ErrCategoryRequired, got: %v", err)
	}
}
