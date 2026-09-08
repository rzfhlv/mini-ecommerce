package infrastructure

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"mini-ecommerce/internal/order/domain"
)

// MidtransPaymentGateway -- Anti-Corruption Layer. Bentuk data
// spesifik Midtrans BERHENTI DI SINI, tidak pernah bocor ke domain.
var (
	_ domain.PaymentGateway = (*MidtransPaymentGateway)(nil)
	_ domain.WebhookParser  = (*MidtransPaymentGateway)(nil)
)

type MidtransPaymentGateway struct {
	baseURL, serverKey string
	httpClient         *http.Client
}

func NewMidtransPaymentGateway(baseURL, serverKey string) *MidtransPaymentGateway {
	return &MidtransPaymentGateway{baseURL: baseURL, serverKey: serverKey, httpClient: &http.Client{}}
}

type midtransChargeRequest struct {
	TransactionDetails struct {
		OrderID     string `json:"order_id"`
		GrossAmount int64  `json:"gross_amount"`
	} `json:"transaction_details"`
}
type midtransChargeResponse struct {
	TransactionID string `json:"transaction_id"`
	RedirectURL   string `json:"redirect_url"`
	StatusCode    string `json:"status_code"`
}

func (g *MidtransPaymentGateway) Initiate(ctx context.Context, orderID uuid.UUID, amount domain.Money) (domain.PaymentInitiation, error) {
	reqBody := midtransChargeRequest{}
	reqBody.TransactionDetails.OrderID = orderID.String()
	reqBody.TransactionDetails.GrossAmount = amount.Amount()
	payload, err := json.Marshal(reqBody)
	if err != nil {
		return domain.PaymentInitiation{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.baseURL+"/v1/charge", bytes.NewReader(payload))
	if err != nil {
		return domain.PaymentInitiation{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(g.serverKey, "")
	resp, err := g.httpClient.Do(req)
	if err != nil {
		return domain.PaymentInitiation{}, fmt.Errorf("payment gateway: %w", err)
	}
	defer resp.Body.Close()
	var result midtransChargeResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return domain.PaymentInitiation{}, err
	}
	return domain.PaymentInitiation{PaymentReference: result.TransactionID, RedirectURL: result.RedirectURL}, nil
}

type midtransWebhookPayload struct {
	OrderID           string `json:"order_id"`
	TransactionID     string `json:"transaction_id"`
	TransactionStatus string `json:"transaction_status"`
}

func (g *MidtransPaymentGateway) ParseWebhook(body []byte) (domain.PaymentNotification, error) {
	var payload midtransWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return domain.PaymentNotification{}, err
	}
	orderID, err := uuid.Parse(payload.OrderID)
	if err != nil {
		return domain.PaymentNotification{}, err
	}
	status := domain.PaymentNotificationFailed
	if payload.TransactionStatus == "settlement" || payload.TransactionStatus == "capture" {
		status = domain.PaymentNotificationPaid
	}
	return domain.PaymentNotification{OrderID: orderID, PaymentReference: payload.TransactionID, Status: status}, nil
}
