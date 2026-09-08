package domain

import (
	"context"

	"github.com/google/uuid"
)

type CartRepository interface {
	FindOrCreateByUserID(ctx context.Context, userID uuid.UUID) (*Cart, error)
	Save(ctx context.Context, cart *Cart) error
}
