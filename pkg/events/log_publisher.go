package events

import (
	"context"
	"log"

	"mini-ecommerce/internal/order/domain"
)

type LogEventPublisher struct{}

func NewLogEventPublisher() *LogEventPublisher { return &LogEventPublisher{} }

func (p *LogEventPublisher) Publish(ctx context.Context, event domain.DomainEvent) {
	log.Printf("[domain-event] %s", event.EventName())
}
