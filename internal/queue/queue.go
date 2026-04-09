package queue

import (
	"context"
	"time"

	"github.com/flp-fernandes/product-views/internal/domain"
)

const (
	batchSize     = 500
	flushInterval = 10 * time.Millisecond
)

type EventQueue struct {
	ch chan domain.ProductView
}

func NewEventQueue(bufferSize int) *EventQueue {
	return &EventQueue{
		ch: make(chan domain.ProductView, bufferSize),
	}
}

func (q *EventQueue) Enqueue(view domain.ProductView) bool {
	select {
	case q.ch <- view:
		return true
	default:
		return false
	}
}

func (q *EventQueue) StartWorkers(
	ctx context.Context,
	n int,
	handler func(context.Context, []domain.ProductView),
) {
	for i := 0; i < n; i++ {
		go func() {
			batch := make([]domain.ProductView, 0, batchSize)
			ticker := time.NewTicker(flushInterval)
			defer ticker.Stop()

			flush := func() {
				if len(batch) > 0 {
					handler(ctx, batch)
					batch = batch[:0]
				}
			}

			for {
				select {
				case <-ctx.Done():
					flush()
					return
				case evt := <-q.ch:
					batch = append(batch, evt)
					if len(batch) >= batchSize {
						flush()
					}
				case <-ticker.C:
					flush()
				}
			}
		}()
	}
}
