package public

import (
	"context"

	"github.com/google/uuid"
	"mini-ecommerce/internal/catalog/domain"
)

// CatalogPublicAPI -- Open Host Service, SATU-SATUNYA pintu context
// lain (Order, Cart) boleh pakai untuk bicara ke Catalog.
type CatalogPublicAPI interface {
	GetStock(ctx context.Context, productID uuid.UUID) (int, error)
	GetPriceAndName(ctx context.Context, productID uuid.UUID) (price int64, name string, err error)
	ReduceStock(ctx context.Context, productID uuid.UUID, qty int) error
}

type catalogPublicAPI struct{ productRepo domain.ProductRepository }

func NewCatalogPublicAPI(productRepo domain.ProductRepository) CatalogPublicAPI { return &catalogPublicAPI{productRepo: productRepo} }

func (a *catalogPublicAPI) GetStock(ctx context.Context, productID uuid.UUID) (int, error) {
	product, err := a.productRepo.FindByID(ctx, productID)
	if err != nil {
		return 0, err
	}
	return product.Stock(), nil
}

func (a *catalogPublicAPI) GetPriceAndName(ctx context.Context, productID uuid.UUID) (int64, string, error) {
	product, err := a.productRepo.FindByID(ctx, productID)
	if err != nil {
		return 0, "", err
	}
	return product.Price().Amount(), product.Name(), nil
}

func (a *catalogPublicAPI) ReduceStock(ctx context.Context, productID uuid.UUID, qty int) error {
	product, err := a.productRepo.FindByID(ctx, productID)
	if err != nil {
		return err
	}
	if err := product.DecreaseStock(qty); err != nil {
		return err
	}
	return a.productRepo.Update(ctx, product)
}
