package kafka

import (
	"context"

	"github.com/segmentio/kafka-go"

	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/config"
)

func NewKafkaConn(cfg *config.Config) (*kafka.Conn, error) {
	return kafka.DialContext(context.Background(), "tcp", cfg.Kafka.Brokers[0])
}
