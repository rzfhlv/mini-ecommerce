package domain

import (
	"context"

	"github.com/google/uuid"
)

type ProductRepository interface {
	Save(ctx context.Context, product *Product) error
	Update(ctx context.Context, product *Product) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*Product, error)
	List(ctx context.Context, categorySlug string, offset, limit int) ([]*Product, int64, error)
}

type CategoryRepository interface {
	List(ctx context.Context) ([]*Category, error)
}
