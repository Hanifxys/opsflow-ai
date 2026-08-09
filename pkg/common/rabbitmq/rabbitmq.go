package rabbitmq

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/opsflow/common/events"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Client struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

type Config struct {
	Host     string
	Port     int
	User     string
	Password string
}

func (c Config) DSN() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%d/", c.User, c.Password, c.Host, c.Port)
}

// NewClient connects to RabbitMQ and declares standard exchanges.
func NewClient(cfg Config) (*Client, error) {
	conn, err := amqp.Dial(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to rabbitmq: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open rabbitmq channel: %w", err)
	}

	// 1. Declare Topic Exchange: opsflow.events
	err = ch.ExchangeDeclare(
		events.ExchangeEvents, // name
		"topic",               // type
		true,                  // durable
		false,                 // auto-deleted
		false,                 // internal
		false,                 // no-wait
		nil,                   // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare exchange %s: %w", events.ExchangeEvents, err)
	}

	// 2. Declare Dead Letter Exchange: opsflow.dlx
	err = ch.ExchangeDeclare(
		events.ExchangeDLX, // name
		"topic",            // type
		true,               // durable
		false,              // auto-deleted
		false,              // internal
		false,              // no-wait
		nil,                // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare DLX %s: %w", events.ExchangeDLX, err)
	}

	return &Client{conn: conn, ch: ch}, nil
}

func (c *Client) Close() {
	if c.ch != nil {
		_ = c.ch.Close()
	}
	if c.conn != nil {
		_ = c.conn.Close()
	}
}

// DeclareQueueWithDLQ declares a primary queue bound to DLX, plus a matching .dlq queue.
func (c *Client) DeclareQueueWithDLQ(queueName, routingKey string) (amqp.Queue, error) {
	dlqName := queueName + ".dlq"

	// 1. Declare DLQ
	_, err := c.ch.QueueDeclare(
		dlqName, true, false, false, false, nil,
	)
	if err != nil {
		return amqp.Queue{}, fmt.Errorf("failed to declare DLQ %s: %w", dlqName, err)
	}

	// Bind DLQ to opsflow.dlx
	err = c.ch.QueueBind(
		dlqName, routingKey+".dlq", events.ExchangeDLX, false, nil,
	)
	if err != nil {
		return amqp.Queue{}, fmt.Errorf("failed to bind DLQ %s: %w", dlqName, err)
	}

	// 2. Declare primary queue with DLX args
	args := amqp.Table{
		"x-dead-letter-exchange":    events.ExchangeDLX,
		"x-dead-letter-routing-key": routingKey + ".dlq",
	}

	q, err := c.ch.QueueDeclare(
		queueName, true, false, false, false, args,
	)
	if err != nil {
		return amqp.Queue{}, fmt.Errorf("failed to declare queue %s: %w", queueName, err)
	}

	// Bind primary queue to opsflow.events
	err = c.ch.QueueBind(
		queueName, routingKey, events.ExchangeEvents, false, nil,
	)
	if err != nil {
		return amqp.Queue{}, fmt.Errorf("failed to bind queue %s: %w", queueName, err)
	}

	return q, nil
}

// Publish Event payload to RabbitMQ.
func (c *Client) Publish(ctx context.Context, routingKey string, body []byte) error {
	pubCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.ch.PublishWithContext(
		pubCtx,
		events.ExchangeEvents, // exchange
		routingKey,            // routing key
		false,                 // mandatory
		false,                 // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now().UTC(),
			Body:         body,
		},
	)
}

// Consume starts consuming messages from the specified queue.
func (c *Client) Consume(queueName, consumerTag string) (<-chan amqp.Delivery, error) {
	return c.ch.Consume(
		queueName,
		consumerTag,
		false, // auto-ack (false = manual ack for idempotency and safety)
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
}

func LogDelivery(logger *slog.Logger, d amqp.Delivery, msg string) {
	logger.Info(msg,
		slog.String("consumer_tag", d.ConsumerTag),
		slog.String("routing_key", d.RoutingKey),
		slog.Uint64("delivery_tag", d.DeliveryTag),
		slog.Bool("redelivered", d.Redelivered),
	)
}
