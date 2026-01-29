package kafka

import (
	"context"
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/compress"

	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/config"
	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/internal/models"
	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/pkg/logger"
)

const (
	writerRequiredAcks = kafka.RequireAll
	writerMaxAttempts  = 3
	writerReadTimeout  = 10 * time.Second
	writerWriteTimeout = 10 * time.Second
)

// OrdersProducer interface
type OrdersProducer interface {
	PublishOrderCreated(ctx context.Context, order *models.Order) error
	PublishOrderUpdated(ctx context.Context, order *models.Order) error
	PublishOrderCancelled(ctx context.Context, order *models.Order) error
	Close()
	Run()
}

type ordersProducer struct {
	log             logger.Logger
	cfg             *config.Config
	createdWriter   *kafka.Writer
	updatedWriter   *kafka.Writer
	cancelledWriter *kafka.Writer
}

// NewOrdersProducer creates a new orders producer
func NewOrdersProducer(log logger.Logger, cfg *config.Config) *ordersProducer {
	return &ordersProducer{log: log, cfg: cfg}
}

// GetNewKafkaWriter creates new kafka writer
func (p *ordersProducer) GetNewKafkaWriter(topic string) *kafka.Writer {
	w := &kafka.Writer{
		Addr:         kafka.TCP(p.cfg.Kafka.Brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: writerRequiredAcks,
		MaxAttempts:  writerMaxAttempts,
		Logger:       kafka.LoggerFunc(p.log.Debugf),
		ErrorLogger:  kafka.LoggerFunc(p.log.Errorf),
		Compression:  compress.Snappy,
		ReadTimeout:  writerReadTimeout,
		WriteTimeout: writerWriteTimeout,
	}
	return w
}

// Run initializes producers writers
func (p *ordersProducer) Run() {
	p.createdWriter = p.GetNewKafkaWriter(OrderCreatedTopic)
	p.updatedWriter = p.GetNewKafkaWriter(OrderUpdatedTopic)
	p.cancelledWriter = p.GetNewKafkaWriter(OrderCancelledTopic)
}

// Close closes writers
func (p *ordersProducer) Close() {
	if p.createdWriter != nil {
		p.createdWriter.Close()
	}
	if p.updatedWriter != nil {
		p.updatedWriter.Close()
	}
	if p.cancelledWriter != nil {
		p.cancelledWriter.Close()
	}
}

// PublishOrderCreated publishes an order created event
func (p *ordersProducer) PublishOrderCreated(ctx context.Context, order *models.Order) error {
	event := models.OrderEvent{
		EventType: "order.created",
		OrderID:   order.OrderID.Hex(),
		UserID:    order.UserID,
		Status:    string(order.Status),
		Timestamp: time.Now().UTC(),
		Data:      order,
	}

	data, err := json.Marshal(event)
	if err != nil {
		p.log.Errorf("ordersProducer.PublishOrderCreated.Marshal: %v", err)
		return err
	}

	msg := kafka.Message{
		Key:   []byte(order.OrderID.Hex()),
		Value: data,
		Time:  time.Now().UTC(),
	}

	return p.createdWriter.WriteMessages(ctx, msg)
}

// PublishOrderUpdated publishes an order updated event
func (p *ordersProducer) PublishOrderUpdated(ctx context.Context, order *models.Order) error {
	event := models.OrderEvent{
		EventType: "order.updated",
		OrderID:   order.OrderID.Hex(),
		UserID:    order.UserID,
		Status:    string(order.Status),
		Timestamp: time.Now().UTC(),
		Data:      order,
	}

	data, err := json.Marshal(event)
	if err != nil {
		p.log.Errorf("ordersProducer.PublishOrderUpdated.Marshal: %v", err)
		return err
	}

	msg := kafka.Message{
		Key:   []byte(order.OrderID.Hex()),
		Value: data,
		Time:  time.Now().UTC(),
	}

	return p.updatedWriter.WriteMessages(ctx, msg)
}

// PublishOrderCancelled publishes an order cancelled event
func (p *ordersProducer) PublishOrderCancelled(ctx context.Context, order *models.Order) error {
	event := models.OrderEvent{
		EventType: "order.cancelled",
		OrderID:   order.OrderID.Hex(),
		UserID:    order.UserID,
		Status:    string(order.Status),
		Timestamp: time.Now().UTC(),
		Data:      order,
	}

	data, err := json.Marshal(event)
	if err != nil {
		p.log.Errorf("ordersProducer.PublishOrderCancelled.Marshal: %v", err)
		return err
	}

	msg := kafka.Message{
		Key:   []byte(order.OrderID.Hex()),
		Value: data,
		Time:  time.Now().UTC(),
	}

	return p.cancelledWriter.WriteMessages(ctx, msg)
}
