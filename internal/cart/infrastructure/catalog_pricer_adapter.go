package infrastructure

import (
	"context"

	"github.com/google/uuid"
	catalogpublic "mini-ecommerce/internal/catalog/public"
)

// CatalogPricerAdapter -- implementasi domain.ProductPricer,
// membungkus CatalogPublicAPI. Pola identik dengan
// order/infrastructure/catalog_stock_checker_adapter.go, beda
// interface port yang dibungkus.
type CatalogPricerAdapter struct{ catalogAPI catalogpublic.CatalogPublicAPI }

func NewCatalogPricerAdapter(catalogAPI catalogpublic.CatalogPublicAPI) *CatalogPricerAdapter { return &CatalogPricerAdapter{catalogAPI: catalogAPI} }

func (a *CatalogPricerAdapter) GetPriceAndName(ctx context.Context, productID uuid.UUID) (int64, string, error) {
	return a.catalogAPI.GetPriceAndName(ctx, productID)
}
