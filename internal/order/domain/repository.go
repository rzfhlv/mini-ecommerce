package domain

import (
	"context"

	"github.com/google/uuid"
)

type OrderRepository interface {
	Save(ctx context.Context, order *Order) error
	Update(ctx context.Context, order *Order) error
	FindByID(ctx context.Context, id uuid.UUID) (*Order, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]*Order, error)
}

// ProductStockChecker -- port milik Order sendiri untuk bicara ke Catalog.
type ProductStockChecker interface {
	GetAvailableStock(ctx context.Context, productID uuid.UUID) (int, error)
	GetProductPrice(ctx context.Context, productID uuid.UUID) (Money, string, error)
	ReduceStock(ctx context.Context, productID uuid.UUID, qty int) error
}
