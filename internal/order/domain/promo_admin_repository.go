package domain

import (
	"context"

	"github.com/google/uuid"
)

// PromoCodeAdminRepository -- TERPISAH dari PromoCodeRepository,
// melayani kebutuhan ADMIN CRUD (bukan checkout). Interface
// Segregation: dua konsumen berbeda, dua interface kecil.
type PromoCodeAdminRepository interface {
	Save(ctx context.Context, promo *PromoCode) error
	Update(ctx context.Context, promo *PromoCode) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*PromoCode, error)
	ExistsByCode(ctx context.Context, code string) (bool, error)
	List(ctx context.Context, offset, limit int) ([]*PromoCode, error)
}
