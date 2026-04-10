package events

import (
	"context"
	"ride-sharing/shared/contracts"
	"ride-sharing/shared/messaging"
)

type TripEventPublisher struct {
	rabbitmq *messaging.RabbitMQ
}

func NewTripEventPublisher(rabbitmq *messaging.RabbitMQ) *TripEventPublisher {
	return &TripEventPublisher{rabbitmq: rabbitmq}
}

func (p *TripEventPublisher) PublishTripCreatedEvent(ctx context.Context) error {
	return p.rabbitmq.Publish(ctx, messaging.TripExchange, contracts.TripEventCreated, []byte(`{"event": "trip_created"}`))
}
