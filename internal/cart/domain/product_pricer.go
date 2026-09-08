package domain

import (
	"context"

	"github.com/google/uuid"
)

// ProductPricer -- port KECIL milik Cart sendiri, terpisah dari
// port milik Order (ProductStockChecker) walau sama-sama konsumen
// Catalog. Setiap context definisikan port sesuai kebutuhannya sendiri.
type ProductPricer interface {
	GetPriceAndName(ctx context.Context, productID uuid.UUID) (price int64, name string, err error)
}
