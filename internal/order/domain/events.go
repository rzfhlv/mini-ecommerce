package domain

import (
	"time"

	"github.com/google/uuid"
)

type DomainEvent interface{ EventName() string }

type OrderPlacedEvent struct {
	OrderID   uuid.UUID
	UserID    uuid.UUID
	Total     Money
	OccuredAt time.Time
}

func (e OrderPlacedEvent) EventName() string { return "order.placed" }

type OrderPaidEvent struct {
	OrderID   uuid.UUID
	OccuredAt time.Time
}

func (e OrderPaidEvent) EventName() string { return "order.paid" }

type OrderDiscountAppliedEvent struct {
	OrderID        uuid.UUID
	DiscountAmount Money
	DiscountLabel  string
}

func (e OrderDiscountAppliedEvent) EventName() string { return "order.discount_applied" }

// OrderAdminFeeAppliedEvent -- BARU, mengikuti pola event diskon.
type OrderAdminFeeAppliedEvent struct {
	OrderID   uuid.UUID
	FeeAmount Money
	FeeLabel  string
}

func (e OrderAdminFeeAppliedEvent) EventName() string { return "order.admin_fee_applied" }
