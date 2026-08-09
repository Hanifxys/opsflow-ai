package amqp

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/opsflow/common/rabbitmq"
	"github.com/opsflow/notification-service/internal/application"
	amqpdriver "github.com/rabbitmq/amqp091-go"
)

type NotificationConsumer struct {
	client  *rabbitmq.Client
	service *application.NotificationService
	logger  *slog.Logger
	stopCh  chan struct{}
	wg      sync.WaitGroup
}

func NewNotificationConsumer(client *rabbitmq.Client, service *application.NotificationService, logger *slog.Logger) *NotificationConsumer {
	return &NotificationConsumer{
		client:  client,
		service: service,
		logger:  logger,
		stopCh:  make(chan struct{}),
	}
}

func (c *NotificationConsumer) Start() error {
	queueName := "notification.worker"
	routingKey := "incident.*" // Subscribe to incident events

	// Declare primary queue + DLQ
	_, err := c.client.DeclareQueueWithDLQ(queueName, routingKey)
	if err != nil {
		return err
	}

	deliveries, err := c.client.Consume(queueName, "notification-worker-1")
	if err != nil {
		return err
	}

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		c.logger.Info("notification worker consumer started", slog.String("queue", queueName))

		for {
			select {
			case <-c.stopCh:
				return
			case d, ok := <-deliveries:
				if !ok {
					c.logger.Info("delivery channel closed")
					return
				}
				c.handleDelivery(d)
			}
		}
	}()

	return nil
}

func (c *NotificationConsumer) Stop() {
	close(c.stopCh)
	c.wg.Wait()
}

func (c *NotificationConsumer) handleDelivery(d amqpdriver.Delivery) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rabbitmq.LogDelivery(c.logger, d, "processing message")

	isNew, err := c.service.ProcessEvent(ctx, "EMAIL", d.RoutingKey, d.Body)
	if err != nil {
		c.logger.Error("failed to process event, sending to DLQ",
			slog.String("routing_key", d.RoutingKey),
			slog.String("error", err.Error()),
		)
		// Reject without requeue -> routes message directly to DLQ
		_ = d.Nack(false, false)
		return
	}

	if isNew {
		c.logger.Info("notification sent & recorded", slog.String("routing_key", d.RoutingKey))
	} else {
		c.logger.Info("duplicate event ignored (idempotent)", slog.String("routing_key", d.RoutingKey))
	}

	// Manual ACK
	_ = d.Ack(false)
}
