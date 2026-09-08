package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"mini-ecommerce/internal/order/domain"
)

var _ domain.OrderRepository = (*PostgresOrderRepository)(nil)

type PostgresOrderRepository struct{ db *sql.DB }

func NewPostgresOrderRepository(db *sql.DB) *PostgresOrderRepository { return &PostgresOrderRepository{db: db} }

func (r *PostgresOrderRepository) Save(ctx context.Context, order *domain.Order) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx, `
		INSERT INTO orders (id, user_id, subtotal, discount_amount, discount_label, admin_fee_amount, admin_fee_label, total_amount, status, payment_reference, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`, order.ID(), order.UserID(), order.Subtotal().Amount(), order.DiscountAmount().Amount(), nullableString(order.DiscountLabel()),
		order.AdminFeeAmount().Amount(), nullableString(order.AdminFeeLabel()), order.TotalAmount().Amount(), order.Status(),
		nullableString(order.PaymentReference()), order.CreatedAt())
	if err != nil {
		return err
	}
	stmt, err := tx.PrepareContext(ctx, `INSERT INTO order_items (id, order_id, product_id, product_name, price, quantity, subtotal) VALUES ($1, $2, $3, $4, $5, $6, $7)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, item := range order.Items() {
		if _, err = stmt.ExecContext(ctx, uuid.New(), order.ID(), item.ProductID(), item.ProductName(), item.Price().Amount(), item.Quantity(), item.Subtotal().Amount()); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *PostgresOrderRepository) Update(ctx context.Context, order *domain.Order) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE orders SET status = $1, payment_reference = $2, paid_at = $3, updated_at = now() WHERE id = $4
	`, order.Status(), nullableString(order.PaymentReference()), order.PaidAt(), order.ID())
	return err
}

func (r *PostgresOrderRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, subtotal, discount_amount, discount_label, admin_fee_amount, admin_fee_label, total_amount, status, payment_reference, created_at, paid_at
		FROM orders WHERE id = $1
	`, id)
	var (
		orderID, userID                                uuid.UUID
		subtotal, discountAmount, adminFeeAmount, total int64
		discountLabel, adminFeeLabel, paymentReference  sql.NullString
		status                                          string
		createdAt                                       time.Time
		paidAt                                          sql.NullTime
	)
	if err := row.Scan(&orderID, &userID, &subtotal, &discountAmount, &discountLabel, &adminFeeAmount, &adminFeeLabel, &total, &status, &paymentReference, &createdAt, &paidAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	items, err := r.findItemsByOrderID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	subtotalMoney, err := domain.NewMoney(subtotal, domain.IDR)
	if err != nil {
		return nil, err
	}
	discountMoney, err := domain.NewMoney(discountAmount, domain.IDR)
	if err != nil {
		return nil, err
	}
	adminFeeMoney, err := domain.NewMoney(adminFeeAmount, domain.IDR)
	if err != nil {
		return nil, err
	}
	totalMoney, err := domain.NewMoney(total, domain.IDR)
	if err != nil {
		return nil, err
	}
	var paidAtPtr *time.Time
	if paidAt.Valid {
		paidAtPtr = &paidAt.Time
	}
	return domain.ReconstructOrder(orderID, userID, items, subtotalMoney, discountMoney, discountLabel.String, adminFeeMoney, adminFeeLabel.String, totalMoney, domain.OrderStatus(status), paymentReference.String, createdAt, paidAtPtr), nil
}

func (r *PostgresOrderRepository) findItemsByOrderID(ctx context.Context, orderID uuid.UUID) ([]domain.OrderItem, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, product_id, product_name, price, quantity FROM order_items WHERE order_id = $1`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []domain.OrderItem
	for rows.Next() {
		var id, productID uuid.UUID
		var name string
		var price int64
		var qty int
		if err := rows.Scan(&id, &productID, &name, &price, &qty); err != nil {
			return nil, err
		}
		priceMoney, err := domain.NewMoney(price, domain.IDR)
		if err != nil {
			return nil, err
		}
		items = append(items, domain.ReconstructOrderItem(id, productID, name, priceMoney, qty))
	}
	return items, rows.Err()
}

func (r *PostgresOrderRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Order, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id FROM orders WHERE user_id = $1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	orders := make([]*domain.Order, 0, len(ids))
	for _, id := range ids {
		order, err := r.FindByID(ctx, id)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	return orders, nil
}

func nullableString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
