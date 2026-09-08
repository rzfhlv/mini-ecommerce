package infrastructure

import (
	"context"

	"github.com/google/uuid"
	catalogpublic "mini-ecommerce/internal/catalog/public"
	"mini-ecommerce/internal/order/domain"
)

// CatalogStockCheckerAdapter -- implementasi domain.ProductStockChecker,
// membungkus CatalogPublicAPI. TIDAK ADA satu baris SQL pun di sini --
// Order tidak pernah tahu Catalog pakai PostgreSQL atau tabel apa.
var _ domain.ProductStockChecker = (*CatalogStockCheckerAdapter)(nil)

type CatalogStockCheckerAdapter struct{ catalogAPI catalogpublic.CatalogPublicAPI }

func NewCatalogStockCheckerAdapter(catalogAPI catalogpublic.CatalogPublicAPI) *CatalogStockCheckerAdapter { return &CatalogStockCheckerAdapter{catalogAPI: catalogAPI} }

func (a *CatalogStockCheckerAdapter) GetAvailableStock(ctx context.Context, productID uuid.UUID) (int, error) {
	return a.catalogAPI.GetStock(ctx, productID)
}
func (a *CatalogStockCheckerAdapter) GetProductPrice(ctx context.Context, productID uuid.UUID) (domain.Money, string, error) {
	price, name, err := a.catalogAPI.GetPriceAndName(ctx, productID)
	if err != nil {
		return domain.Money{}, "", err
	}
	money, err := domain.NewMoney(price, domain.IDR)
	if err != nil {
		return domain.Money{}, "", err
	}
	return money, name, nil
}
func (a *CatalogStockCheckerAdapter) ReduceStock(ctx context.Context, productID uuid.UUID, qty int) error {
	return a.catalogAPI.ReduceStock(ctx, productID, qty)
}
