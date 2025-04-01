package services

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

type EventType string

const (
	EventTypeMessageSent     EventType = "message.sent"
	EventTypeMessageReceived EventType = "message.received"
	EventTypeConnectionState EventType = "connection.state"
)

type Event struct {
	Type      EventType   `json:"type"`
	CompanyID uuid.UUID   `json:"company_id"`
	UserID    uuid.UUID   `json:"user_id"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

type EventService struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	mu      sync.RWMutex
}

var (
	eventService *EventService
	eventOnce    sync.Once
)

// GetEventService returns the singleton instance of EventService
func GetEventService() (*EventService, error) {
	var err error
	eventOnce.Do(func() {
		eventService, err = newEventService()
	})
	return eventService, err
}

// newEventService creates a new instance of EventService
func newEventService() (*EventService, error) {
	// Connect to RabbitMQ
	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%s%s",
		config.RabbitMQUsername,
		config.RabbitMQPassword,
		config.RabbitMQHost,
		config.RabbitMQPort,
		config.RabbitMQVHost,
	))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %v", err)
	}

	// Create a channel
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %v", err)
	}

	// Declare the exchange
	err = ch.ExchangeDeclare(
		"whatsapp_events", // name
		"topic",           // type
		true,              // durable
		false,             // auto-deleted
		false,             // internal
		false,             // no-wait
		nil,               // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare exchange: %v", err)
	}

	return &EventService{
		conn:    conn,
		channel: ch,
	}, nil
}

// PublishEvent publishes an event to RabbitMQ
func (s *EventService) PublishEvent(ctx context.Context, event Event) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Marshal the event
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %v", err)
	}

	// Publish the event
	err = s.channel.PublishWithContext(ctx,
		"whatsapp_events",  // exchange
		string(event.Type), // routing key
		false,              // mandatory
		false,              // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
	if err != nil {
		return fmt.Errorf("failed to publish event: %v", err)
	}

	return nil
}

// SubscribeToEvents subscribes to events of a specific type
func (s *EventService) SubscribeToEvents(eventType EventType, handler func(Event) error) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Declare a queue
	q, err := s.channel.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %v", err)
	}

	// Bind the queue to the exchange
	err = s.channel.QueueBind(
		q.Name,            // queue name
		string(eventType), // routing key
		"whatsapp_events", // exchange
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue: %v", err)
	}

	// Start consuming messages
	msgs, err := s.channel.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return fmt.Errorf("failed to register a consumer: %v", err)
	}

	// Start a goroutine to process messages
	go func() {
		for msg := range msgs {
			var event Event
			if err := json.Unmarshal(msg.Body, &event); err != nil {
				fmt.Printf("Failed to unmarshal event: %v\n", err)
				continue
			}

			if err := handler(event); err != nil {
				fmt.Printf("Failed to handle event: %v\n", err)
			}
		}
	}()

	return nil
}

// Close closes the RabbitMQ connection
func (s *EventService) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.channel.Close(); err != nil {
		return fmt.Errorf("failed to close channel: %v", err)
	}

	if err := s.conn.Close(); err != nil {
		return fmt.Errorf("failed to close connection: %v", err)
	}

	return nil
}
