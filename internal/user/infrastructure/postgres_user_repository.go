package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"mini-ecommerce/internal/user/domain"
)

type scanRow interface{ Scan(dest ...interface{}) error }

type PostgresUserRepository struct{ db *sql.DB }

var _ domain.UserRepository = (*PostgresUserRepository)(nil)

func NewPostgresUserRepository(db *sql.DB) *PostgresUserRepository { return &PostgresUserRepository{db: db} }

func (r *PostgresUserRepository) Save(ctx context.Context, user *domain.User) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO users (id, name, email, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $5)
	`, user.ID(), user.Name(), user.Email().String(), user.PasswordHash().String(), user.CreatedAt())
	return err
}

func (r *PostgresUserRepository) FindByEmail(ctx context.Context, email domain.Email) (*domain.User, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, name, email, password_hash, created_at, updated_at FROM users WHERE email = $1`, email.String())
	return scanUser(row)
}

func (r *PostgresUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, name, email, password_hash, created_at, updated_at FROM users WHERE id = $1`, id)
	return scanUser(row)
}

func (r *PostgresUserRepository) ExistsByEmail(ctx context.Context, email domain.Email) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`, email.String()).Scan(&exists)
	return exists, err
}

func scanUser(row scanRow) (*domain.User, error) {
	var (
		id                     uuid.UUID
		name, emailStr, hashed string
		createdAt, updatedAt   time.Time
	)
	if err := row.Scan(&id, &name, &emailStr, &hashed, &createdAt, &updatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, err
	}
	email, err := domain.NewEmail(emailStr)
	if err != nil {
		return nil, err
	}
	passwordHash, err := domain.NewPasswordHash(hashed)
	if err != nil {
		return nil, err
	}
	return domain.ReconstructUser(id, name, email, passwordHash, createdAt, updatedAt), nil
}
