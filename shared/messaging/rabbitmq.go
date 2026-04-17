package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"ride-sharing/shared/contracts"
	"ride-sharing/shared/retry"
	"ride-sharing/shared/tracing"

	amqp "github.com/rabbitmq/amqp091-go"
)

const TripExchange = "x.trip"

type RabbitMQ struct {
	conn    *amqp.Connection
	Channel *amqp.Channel
}

func NewRabbitMQ(url string) (*RabbitMQ, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	return &RabbitMQ{conn: conn, Channel: ch}, nil
}

func (r *RabbitMQ) SetupExchangesAndQueues() error {
	// Example: Declare an exchange and a queue, and bind them
	err := r.Channel.ExchangeDeclare(
		TripExchange, // name
		"topic",      // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	err = r.declareAndBindQueue(
		FindAvailableDriversQueue,
		TripExchange,
		[]string{
			contracts.TripEventCreated,
			contracts.TripEventDriverNotInterested,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to declare and bind queue: %w", err)
	}

	err = r.declareAndBindQueue(
		DriverCmdTripRequestQueue,
		TripExchange,
		[]string{
			contracts.DriverCmdTripRequest,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to declare and bind queue: %w", err)
	}

	err = r.declareAndBindQueue(
		DriverTripResponseQueue,
		TripExchange,
		[]string{
			contracts.DriverCmdTripAccept,
			contracts.DriverCmdTripDecline,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to declare and bind queue: %w", err)
	}

	err = r.declareAndBindQueue(
		NotifyDriverNotFoundQueue,
		TripExchange,
		[]string{
			contracts.TripEventNoDriversFound,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to declare and bind queue: %w", err)
	}

	if err := r.declareAndBindQueue(
		NotifyDriverAssignQueue,
		TripExchange,
		[]string{contracts.TripEventDriverAssigned},
	); err != nil {
		return fmt.Errorf("failed to declare and bind queue: %w", err)
	}

	if err := r.declareAndBindQueue(
		PaymentTripResponseQueue,
		TripExchange,
		[]string{contracts.PaymentCmdCreateSession},
	); err != nil {
		return fmt.Errorf("failed to declare and bind queue: %w", err)
	}

	if err := r.declareAndBindQueue(
		NotifyPaymentSessionCreatedQueue,
		TripExchange,
		[]string{contracts.PaymentEventSessionCreated},
	); err != nil {
		return fmt.Errorf("failed to declare and bind queue: %w", err)
	}

	if err := r.declareAndBindQueue(
		NotifyPaymentSuccessQueue,
		TripExchange,
		[]string{contracts.PaymentEventSuccess},
	); err != nil {
		return fmt.Errorf("failed to declare and bind queue: %w", err)
	}

	// Additional queues and bindings can be set up here

	return nil
}

func (r *RabbitMQ) declareAndBindQueue(queueName, exchangeName string, messageTypes []string) error {
	_, err := r.Channel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	for _, msgType := range messageTypes {
		err = r.Channel.QueueBind(
			queueName,    // queue name
			msgType,      // routing key
			exchangeName, // exchange
			false,
			nil,
		)
		if err != nil {
			return fmt.Errorf("failed to bind queue: %w", err)
		}
	}

	return nil
}

func (r *RabbitMQ) Publish(ctx context.Context, exchange, routingKey string, msg contracts.AmqpMessage) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	publishing := amqp.Publishing{
		ContentType:  "text/plain",
		Body:         body,
		DeliveryMode: amqp.Persistent,
	}

	return tracing.TracedPublisher(
		ctx,
		exchange,
		routingKey,
		publishing,
		func(publishCtx context.Context, publishExchange, publishRoutingKey string, publishMsg amqp.Publishing) error {
			return r.Channel.PublishWithContext(
				publishCtx,
				publishExchange,
				publishRoutingKey,
				false, // mandatory
				false, // immediate
				publishMsg,
			)
		},
	)
}

type MessageHandler func(context.Context, amqp.Delivery) error

func (r *RabbitMQ) ConsumeMessages(queueName string, handler MessageHandler) error {
	err := r.Channel.Qos(1, 0, false)
	if err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	msgs, err := r.Channel.Consume(
		queueName, // queue
		"",        // consumer
		false,     // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		return err
	}

	go func() {
		for msg := range msgs {
			if err := tracing.TracedConsumer(msg, func(ctx context.Context, d amqp.Delivery) error {
				log.Printf("Received a message: %s", msg.Body)

				cfg := retry.DefaultConfig()
				err := retry.WithBackoff(ctx, cfg, func() error {
					return handler(ctx, d)
				})
				if err != nil {
					log.Printf("Message processing failed after %d retries for message ID: %s, err: %v", cfg.MaxRetries, d.MessageId, err)

					// Add failure context before sending to the DLQ
					headers := amqp.Table{}
					if d.Headers != nil {
						headers = d.Headers
					}

					headers["x-death-reason"] = err.Error()
					headers["x-origin-exchange"] = d.Exchange
					headers["x-original-routing-key"] = d.RoutingKey
					headers["x-retry-count"] = cfg.MaxRetries
					d.Headers = headers

					// Reject without requeue - message will go to the DLQ
					_ = d.Reject(false)
					return err
				}

				// Only Ack if the handler succeeds
				if ackErr := msg.Ack(false); ackErr != nil {
					log.Printf("ERROR: Failed to Ack message: %v. Message body: %s", ackErr, msg.Body)
				}

				return nil
			}); err != nil {
				log.Printf("Error processing message: %v", err)
			}
		}
	}()

	return nil
}

func (r *RabbitMQ) Close() error {
	if r.conn != nil {
		if err := r.conn.Close(); err != nil {
			return fmt.Errorf("failed to close RabbitMQ connection: %w", err)
		}
	}
	if r.Channel != nil {
		if err := r.Channel.Close(); err != nil {
			return fmt.Errorf("failed to close RabbitMQ channel: %w", err)
		}
	}
	return nil
}
