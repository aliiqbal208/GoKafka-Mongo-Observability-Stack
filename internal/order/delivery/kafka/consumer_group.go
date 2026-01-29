package kafka

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/compress"

	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/config"
	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/internal/models"
	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/internal/order"
	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/internal/product"
	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/pkg/logger"
)

const (
	minBytes               = 10e3 // 10KB
	maxBytes               = 10e6 // 10MB
	queueCapacity          = 100
	heartbeatInterval      = 3 * time.Second
	commitInterval         = 0
	partitionWatchInterval = 5 * time.Second
	maxAttempts            = 3
	dialTimeout            = 3 * time.Minute
)

// OrdersConsumerGroup struct
type OrdersConsumerGroup struct {
	Brokers     []string
	GroupID     string
	log         logger.Logger
	cfg         *config.Config
	orderUC     order.UseCase
	productRepo product.MongoRepository
	validate    *validator.Validate
}

// NewOrdersConsumerGroup constructor
func NewOrdersConsumerGroup(
	brokers []string,
	groupID string,
	log logger.Logger,
	cfg *config.Config,
	orderUC order.UseCase,
	productRepo product.MongoRepository,
	validate *validator.Validate,
) *OrdersConsumerGroup {
	return &OrdersConsumerGroup{
		Brokers:     brokers,
		GroupID:     groupID,
		log:         log,
		cfg:         cfg,
		orderUC:     orderUC,
		productRepo: productRepo,
		validate:    validate,
	}
}

func (ocg *OrdersConsumerGroup) getNewKafkaReader(kafkaURL []string, topic, groupID string) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:                kafkaURL,
		GroupID:                groupID,
		Topic:                  topic,
		MinBytes:               minBytes,
		MaxBytes:               maxBytes,
		QueueCapacity:          queueCapacity,
		HeartbeatInterval:      heartbeatInterval,
		CommitInterval:         commitInterval,
		PartitionWatchInterval: partitionWatchInterval,
		Logger:                 kafka.LoggerFunc(ocg.log.Debugf),
		ErrorLogger:            kafka.LoggerFunc(ocg.log.Errorf),
		MaxAttempts:            maxAttempts,
		Dialer: &kafka.Dialer{
			Timeout: dialTimeout,
		},
	})
}

func (ocg *OrdersConsumerGroup) getNewKafkaWriter(topic string) *kafka.Writer {
	w := &kafka.Writer{
		Addr:         kafka.TCP(ocg.Brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: writerRequiredAcks,
		MaxAttempts:  writerMaxAttempts,
		Logger:       kafka.LoggerFunc(ocg.log.Debugf),
		ErrorLogger:  kafka.LoggerFunc(ocg.log.Errorf),
		Compression:  compress.Snappy,
		ReadTimeout:  writerReadTimeout,
		WriteTimeout: writerWriteTimeout,
	}
	return w
}

func (ocg *OrdersConsumerGroup) consumeOrderCreated(
	ctx context.Context,
	cancel context.CancelFunc,
	groupID string,
	topic string,
	workersNum int,
) {
	r := ocg.getNewKafkaReader(ocg.Brokers, topic, groupID)

	defer func() {
		if err := r.Close(); err != nil {
			ocg.log.Errorf("r.Close: %v", err)
			cancel()
		}
	}()

	ocg.log.Infof("Starting consumer group: %s, topic: %s", groupID, topic)

	wg := &sync.WaitGroup{}
	for i := 0; i < workersNum; i++ {
		wg.Add(1)
		go ocg.orderCreatedWorker(ctx, cancel, r, wg, i)
	}
	wg.Wait()
}

func (ocg *OrdersConsumerGroup) consumeOrderUpdated(
	ctx context.Context,
	cancel context.CancelFunc,
	groupID string,
	topic string,
	workersNum int,
) {
	r := ocg.getNewKafkaReader(ocg.Brokers, topic, groupID)

	defer func() {
		if err := r.Close(); err != nil {
			ocg.log.Errorf("r.Close: %v", err)
			cancel()
		}
	}()

	ocg.log.Infof("Starting consumer group: %s, topic: %s", groupID, topic)

	wg := &sync.WaitGroup{}
	for i := 0; i < workersNum; i++ {
		wg.Add(1)
		go ocg.orderUpdatedWorker(ctx, cancel, r, wg, i)
	}
	wg.Wait()
}

// RunConsumers runs all order consumers
func (ocg *OrdersConsumerGroup) RunConsumers(ctx context.Context, cancel context.CancelFunc) {
	go ocg.consumeOrderCreated(ctx, cancel, OrderConsumerGroup, OrderCreatedTopic, CreateOrderWorkers)
	go ocg.consumeOrderUpdated(ctx, cancel, OrderConsumerGroup, OrderUpdatedTopic, UpdateOrderWorkers)
}

// ProcessOrderEvent processes an order event from Kafka
func (ocg *OrdersConsumerGroup) ProcessOrderEvent(ctx context.Context, msg kafka.Message) error {
	var event models.OrderEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		ocg.log.Errorf("OrdersConsumerGroup.ProcessOrderEvent.Unmarshal: %v", err)
		return err
	}

	ocg.log.Infof("Processing order event: type=%s, orderId=%s, status=%s",
		event.EventType, event.OrderID, event.Status)

	switch event.EventType {
	case "order.created":
		return ocg.processOrderCreated(ctx, &event)
	case "order.updated":
		return ocg.processOrderUpdated(ctx, &event)
	case "order.cancelled":
		return ocg.processOrderCancelled(ctx, &event)
	default:
		ocg.log.Warnf("Unknown event type: %s", event.EventType)
	}

	return nil
}

func (ocg *OrdersConsumerGroup) processOrderCreated(ctx context.Context, event *models.OrderEvent) error {
	ocg.log.Infof("Order created: %s for user: %s", event.OrderID, event.UserID)
	
	// Deduct inventory for each item in the order
	if event.Data != nil {
		for _, item := range event.Data.Items {
			// Here you would update the product stock
			ocg.log.Infof("Reserving stock for product %s, quantity: %d", item.ProductID.Hex(), item.Quantity)
		}
	}
	
	return nil
}

func (ocg *OrdersConsumerGroup) processOrderUpdated(ctx context.Context, event *models.OrderEvent) error {
	ocg.log.Infof("Order updated: %s, new status: %s", event.OrderID, event.Status)
	
	// Handle specific status updates
	switch models.OrderStatus(event.Status) {
	case models.OrderStatusConfirmed:
		ocg.log.Infof("Order %s confirmed, preparing for processing", event.OrderID)
	case models.OrderStatusShipped:
		ocg.log.Infof("Order %s shipped, sending notification", event.OrderID)
	case models.OrderStatusDelivered:
		ocg.log.Infof("Order %s delivered", event.OrderID)
	}
	
	return nil
}

func (ocg *OrdersConsumerGroup) processOrderCancelled(ctx context.Context, event *models.OrderEvent) error {
	ocg.log.Infof("Order cancelled: %s", event.OrderID)
	
	// Restore inventory for each item in the order
	if event.Data != nil {
		for _, item := range event.Data.Items {
			ocg.log.Infof("Restoring stock for product %s, quantity: %d", item.ProductID.Hex(), item.Quantity)
		}
	}
	
	return nil
}
