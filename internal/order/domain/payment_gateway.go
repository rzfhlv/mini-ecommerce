package domain

import (
	"context"

	"github.com/google/uuid"
)

type PaymentInitiation struct {
	PaymentReference string
	RedirectURL      string
}

// PaymentGateway -- port arah KELUAR (Order -> gateway eksternal).
type PaymentGateway interface {
	Initiate(ctx context.Context, orderID uuid.UUID, amount Money) (PaymentInitiation, error)
}

// WebhookParser -- port TERPISAH untuk arah MASUK (gateway -> Order),
// dipakai delivery layer, bukan use case. Interface Segregation:
// tiap konsumen cuma lihat method yang dia butuhkan.
type WebhookParser interface {
	ParseWebhook(body []byte) (PaymentNotification, error)
}

type PaymentNotificationStatus string

const (
	PaymentNotificationPaid   PaymentNotificationStatus = "PAID"
	PaymentNotificationFailed PaymentNotificationStatus = "FAILED"
)

type PaymentNotification struct {
	OrderID          uuid.UUID
	PaymentReference string
	Status           PaymentNotificationStatus
}
