package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"airline-booking-system/internal/config"
	"airline-booking-system/internal/models"

	"github.com/segmentio/kafka-go"
)

// Producer handles Kafka message production
type Producer struct {
	writer *kafka.Writer
}

// NewProducer creates a new Kafka producer
func NewProducer(cfg *config.KafkaConfig) *Producer {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(cfg.Brokers...),
		Balancer: &kafka.LeastBytes{},
	}

	return &Producer{writer: writer}
}

// SendPaymentEvent sends a payment event to Kafka
func (p *Producer) SendPaymentEvent(ctx context.Context, event *models.PaymentEvent) error {
	eventData, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal payment event: %w", err)
	}

	message := kafka.Message{
		Topic: "payment-events",
		Key:   []byte(fmt.Sprintf("%d", event.BookingID)),
		Value: eventData,
	}

	err = p.writer.WriteMessages(ctx, message)
	if err != nil {
		return fmt.Errorf("failed to send payment event: %w", err)
	}

	return nil
}

// SendSeatUpdateEvent sends a seat update event to Kafka
func (p *Producer) SendSeatUpdateEvent(ctx context.Context, event *models.SeatUpdateEvent) error {
	eventData, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal seat update event: %w", err)
	}

	message := kafka.Message{
		Topic: "flight-bookings",
		Key:   []byte(fmt.Sprintf("%d", event.FlightID)),
		Value: eventData,
	}

	err = p.writer.WriteMessages(ctx, message)
	if err != nil {
		return fmt.Errorf("failed to send seat update event: %w", err)
	}

	return nil
}

// Close closes the producer
func (p *Producer) Close() error {
	return p.writer.Close()
}
