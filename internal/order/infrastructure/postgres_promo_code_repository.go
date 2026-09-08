package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	"mini-ecommerce/internal/order/domain"
)

var _ domain.PromoCodeRepository = (*PostgresPromoCodeRepository)(nil)

type PostgresPromoCodeRepository struct{ db *sql.DB }

func NewPostgresPromoCodeRepository(db *sql.DB) *PostgresPromoCodeRepository { return &PostgresPromoCodeRepository{db: db} }

// FindPolicy -- BUKAN switch discount_type lagi. Cukup baca row jadi
// domain.PromoCode, lalu serahkan ke PromoCode.ToDiscountPolicy() --
// aturan "tipe -> strategy" cuma hidup di SATU tempat (domain layer).
func (r *PostgresPromoCodeRepository) FindPolicy(ctx context.Context, code string) (domain.DiscountPolicy, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, code, label, discount_type, discount_value, min_purchase, is_active, expires_at, created_at, updated_at FROM promo_codes WHERE code = $1`, code)
	promo, err := scanPromoCode(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrPromoCodeNotFound
		}
		return nil, err
	}
	if !promo.IsUsable() {
		return nil, domain.ErrPromoCodeNotFound
	}
	return promo.ToDiscountPolicy()
}
