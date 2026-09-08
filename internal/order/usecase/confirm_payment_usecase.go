package usecase

import (
	"context"

	"mini-ecommerce/internal/order/domain"
)

type ConfirmPaymentUseCase interface {
	Execute(ctx context.Context, notification domain.PaymentNotification) error
}

type confirmPaymentUseCase struct {
	orderRepo      domain.OrderRepository
	eventPublisher EventPublisher
}

func NewConfirmPaymentUseCase(orderRepo domain.OrderRepository, publisher EventPublisher) ConfirmPaymentUseCase {
	return &confirmPaymentUseCase{orderRepo: orderRepo, eventPublisher: publisher}
}

func (uc *confirmPaymentUseCase) Execute(ctx context.Context, notification domain.PaymentNotification) error {
	order, err := uc.orderRepo.FindByID(ctx, notification.OrderID)
	if err != nil {
		return err
	}
	if order == nil {
		return ErrOrderNotFound
	}
	if notification.Status != domain.PaymentNotificationPaid {
		return nil
	}
	if err := order.MarkAsPaid(); err != nil {
		return err
	}
	if err := uc.orderRepo.Update(ctx, order); err != nil {
		return err
	}
	for _, event := range order.PullEvents() {
		uc.eventPublisher.Publish(ctx, event)
	}
	return nil
}
