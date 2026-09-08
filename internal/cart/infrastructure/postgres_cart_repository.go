package infrastructure

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"mini-ecommerce/internal/cart/domain"
)

var _ domain.CartRepository = (*PostgresCartRepository)(nil)

type PostgresCartRepository struct{ db *sql.DB }

func NewPostgresCartRepository(db *sql.DB) *PostgresCartRepository { return &PostgresCartRepository{db: db} }

func (r *PostgresCartRepository) FindOrCreateByUserID(ctx context.Context, userID uuid.UUID) (*domain.Cart, error) {
	var cartID uuid.UUID
	err := r.db.QueryRowContext(ctx, `SELECT id FROM carts WHERE user_id = $1`, userID).Scan(&cartID)
	if err == sql.ErrNoRows {
		cart := domain.NewCart(userID)
		_, insertErr := r.db.ExecContext(ctx, `INSERT INTO carts (id, user_id, created_at, updated_at) VALUES ($1, $2, now(), now())`, cart.ID(), userID)
		if insertErr != nil {
			return nil, insertErr
		}
		return cart, nil
	}
	if err != nil {
		return nil, err
	}
	rows, err := r.db.QueryContext(ctx, `SELECT product_id, product_name, price_snapshot, quantity FROM cart_items WHERE cart_id = $1`, cartID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []domain.CartItem
	for rows.Next() {
		var productID uuid.UUID
		var productName string
		var price int64
		var qty int
		if err := rows.Scan(&productID, &productName, &price, &qty); err != nil {
			return nil, err
		}
		items = append(items, domain.ReconstructCartItem(productID, productName, price, qty))
	}
	return domain.ReconstructCart(cartID, userID, items, time.Now()), rows.Err()
}

func (r *PostgresCartRepository) Save(ctx context.Context, cart *domain.Cart) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `DELETE FROM cart_items WHERE cart_id = $1`, cart.ID()); err != nil {
		return err
	}
	for _, item := range cart.Items() {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO cart_items (id, cart_id, product_id, product_name, quantity, price_snapshot)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, uuid.New(), cart.ID(), item.ProductID(), item.ProductName(), item.Quantity(), item.PriceSnapshot())
		if err != nil {
			return err
		}
	}
	if _, err := tx.ExecContext(ctx, `UPDATE carts SET updated_at = now() WHERE id = $1`, cart.ID()); err != nil {
		return err
	}
	return tx.Commit()
}
