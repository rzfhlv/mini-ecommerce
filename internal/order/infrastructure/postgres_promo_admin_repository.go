package infrastructure

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"mini-ecommerce/internal/order/domain"
)

var _ domain.PromoCodeAdminRepository = (*PostgresPromoAdminRepository)(nil)

type scanRow interface{ Scan(dest ...interface{}) error }

type PostgresPromoAdminRepository struct{ db *sql.DB }

func NewPostgresPromoAdminRepository(db *sql.DB) *PostgresPromoAdminRepository { return &PostgresPromoAdminRepository{db: db} }

func (r *PostgresPromoAdminRepository) Save(ctx context.Context, promo *domain.PromoCode) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO promo_codes (id, code, label, discount_type, discount_value, min_purchase, is_active, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, promo.ID(), promo.Code(), promo.Label(), string(promo.DiscountType()), promo.DiscountValue(), promo.MinPurchase(), promo.IsActive(), promo.ExpiresAt(), promo.CreatedAt())
	return err
}

func (r *PostgresPromoAdminRepository) Update(ctx context.Context, promo *domain.PromoCode) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE promo_codes SET label = $1, discount_value = $2, min_purchase = $3, is_active = $4, expires_at = $5 WHERE id = $6
	`, promo.Label(), promo.DiscountValue(), promo.MinPurchase(), promo.IsActive(), promo.ExpiresAt(), promo.ID())
	return err
}

func (r *PostgresPromoAdminRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM promo_codes WHERE id = $1`, id)
	return err
}

func (r *PostgresPromoAdminRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.PromoCode, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, code, label, discount_type, discount_value, min_purchase, is_active, expires_at, created_at, updated_at FROM promo_codes WHERE id = $1`, id)
	return scanPromoCode(row)
}

func (r *PostgresPromoAdminRepository) ExistsByCode(ctx context.Context, code string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM promo_codes WHERE code = $1)`, code).Scan(&exists)
	return exists, err
}

func (r *PostgresPromoAdminRepository) List(ctx context.Context, offset, limit int) ([]*domain.PromoCode, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, code, label, discount_type, discount_value, min_purchase, is_active, expires_at, created_at, updated_at
		FROM promo_codes ORDER BY created_at DESC LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*domain.PromoCode
	for rows.Next() {
		promo, err := scanPromoCode(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, promo)
	}
	return result, rows.Err()
}

func scanPromoCode(row scanRow) (*domain.PromoCode, error) {
	var (
		id                    uuid.UUID
		code, label, discType string
		discountValue         int
		minPurchase           sql.NullInt64
		isActive              bool
		expiresAt             sql.NullTime
		createdAt, updatedAt  sql.NullTime
	)
	if err := row.Scan(&id, &code, &label, &discType, &discountValue, &minPurchase, &isActive, &expiresAt, &createdAt, &updatedAt); err != nil {
		return nil, err
	}
	var minPurchasePtr *int64
	if minPurchase.Valid {
		minPurchasePtr = &minPurchase.Int64
	}
	var expiresAtPtr *time.Time
	if expiresAt.Valid {
		expiresAtPtr = &expiresAt.Time
	}
	return domain.ReconstructPromoCode(id, code, label, domain.DiscountType(discType), discountValue, minPurchasePtr, isActive, expiresAtPtr, createdAt.Time, updatedAt.Time), nil
}
