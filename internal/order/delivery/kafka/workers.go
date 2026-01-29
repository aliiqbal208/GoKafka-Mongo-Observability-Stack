package kafka

import (
	"context"
	"sync"

	"github.com/segmentio/kafka-go"
)

func (ocg *OrdersConsumerGroup) orderCreatedWorker(
	ctx context.Context,
	cancel context.CancelFunc,
	r *kafka.Reader,
	wg *sync.WaitGroup,
	workerID int,
) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		m, err := r.FetchMessage(ctx)
		if err != nil {
			ocg.log.Warnf("orderCreatedWorker.FetchMessage: %v", err)
			continue
		}

		ocg.log.Infof(
			"WORKER: %d, message at topic/partition/offset %s/%d/%d: %s = %s\n",
			workerID,
			m.Topic,
			m.Partition,
			m.Offset,
			string(m.Key),
			string(m.Value),
		)

		if err := ocg.ProcessOrderEvent(ctx, m); err != nil {
			ocg.log.Errorf("orderCreatedWorker.ProcessOrderEvent: %v", err)
			continue
		}

		if err := r.CommitMessages(ctx, m); err != nil {
			ocg.log.Errorf("orderCreatedWorker.CommitMessages: %v", err)
		}
	}
}

func (ocg *OrdersConsumerGroup) orderUpdatedWorker(
	ctx context.Context,
	cancel context.CancelFunc,
	r *kafka.Reader,
	wg *sync.WaitGroup,
	workerID int,
) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		m, err := r.FetchMessage(ctx)
		if err != nil {
			ocg.log.Warnf("orderUpdatedWorker.FetchMessage: %v", err)
			continue
		}

		ocg.log.Infof(
			"WORKER: %d, message at topic/partition/offset %s/%d/%d: %s = %s\n",
			workerID,
			m.Topic,
			m.Partition,
			m.Offset,
			string(m.Key),
			string(m.Value),
		)

		if err := ocg.ProcessOrderEvent(ctx, m); err != nil {
			ocg.log.Errorf("orderUpdatedWorker.ProcessOrderEvent: %v", err)
			continue
		}

		if err := r.CommitMessages(ctx, m); err != nil {
			ocg.log.Errorf("orderUpdatedWorker.CommitMessages: %v", err)
		}
	}
}
