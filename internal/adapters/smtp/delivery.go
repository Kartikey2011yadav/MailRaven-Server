package smtp

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/Kartikey2011yadav/mailraven-server/internal/core/domain"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/ports"
	"github.com/Kartikey2011yadav/mailraven-server/internal/observability"
)

// DeliveryWorker processes the outbound email queue
type DeliveryWorker struct {
	queueRepo ports.QueueRepository
	blobStore ports.BlobStore
	sender    Sender
	logger    *observability.Logger
	metrics   *observability.Metrics
	stopChan  chan struct{}
	wg        sync.WaitGroup
	ticker    *time.Ticker
}

// NewDeliveryWorker creates a new delivery worker
func NewDeliveryWorker(
	queueRepo ports.QueueRepository,
	blobStore ports.BlobStore,
	sender Sender,
	logger *observability.Logger,
	metrics *observability.Metrics,
) *DeliveryWorker {
	return &DeliveryWorker{
		queueRepo: queueRepo,
		blobStore: blobStore,
		sender:    sender,
		logger:    logger,
		metrics:   metrics,
		stopChan:  make(chan struct{}),
		ticker:    time.NewTicker(5 * time.Second), // Check queue every 5 seconds
	}
}

// Start begins processing the queue in a background routine
func (w *DeliveryWorker) Start() {
	w.logger.Info("delivery worker started")
	w.wg.Add(1)
	go w.processLoop()
}

// Stop gracefully stops the worker, waiting for current delivery to finish
func (w *DeliveryWorker) Stop() {
	w.logger.Info("stopping delivery worker...")
	close(w.stopChan)
	w.wg.Wait()
	w.ticker.Stop()
	w.logger.Info("delivery worker stopped")
}

func (w *DeliveryWorker) processLoop() {
	defer w.wg.Done()

	for {
		select {
		case <-w.stopChan:
			return
		case <-w.ticker.C:
			w.ProcessNext()
		}
	}
}

func (w *DeliveryWorker) ProcessNext() {
	ctx := context.Background()

	// 1. Lock next ready message
	msg, err := w.queueRepo.LockNextReady(ctx)
	if err != nil {
		w.logger.Error("failed to lock next message", "error", err)
		return
	}
	if msg == nil {
		// Queue empty or no ready messages
		return
	}

	w.logger.Info("proccessing outbound message", "id", msg.ID, "retry", msg.RetryCount)

	// 2. Fetch message content from blob store
	content, err := w.blobStore.Read(ctx, msg.BlobKey)
	if err != nil {
		w.handlePermanentFailure(ctx, msg, fmt.Sprintf("blob missing: %v", err))
		return
	}
	// content is []byte

	// 3. Attempt Delivery
	// Use short timeout for delivery attempt
	deliverCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	start := time.Now()
	err = w.sender.Send(deliverCtx, msg.Sender, msg.Recipient, content)
	duration := time.Since(start)

	if err != nil {
		w.logger.Warn("delivery failed", "id", msg.ID, "error", err, "duration", duration)
		w.handleRetry(ctx, msg, err)
	} else {
		w.logger.Info("delivery succeeded", "id", msg.ID, "duration", duration)
		w.handleSuccess(ctx, msg)
	}
}

func (w *DeliveryWorker) handleSuccess(ctx context.Context, msg *domain.OutboundMessage) {
	if err := w.queueRepo.UpdateStatus(ctx, msg.ID, domain.QueueStatusSent, msg.RetryCount, time.Time{}, ""); err != nil {
		w.logger.Error("failed to mark message as sent", "id", msg.ID, "error", err)
	}
	w.metrics.IncrementOutboundSent()
}

func (w *DeliveryWorker) handleRetry(ctx context.Context, msg *domain.OutboundMessage, failureErr error) {
	// Calculate backoff
	// Strategy: 1min, 5min, 15min, 1h, 6h, 12h, 24h
	retryCount := msg.RetryCount + 1

	// Max retries
	const maxRetries = 10
	if retryCount > maxRetries {
		w.handlePermanentFailure(ctx, msg, fmt.Sprintf("max retries exceeded. last error: %v", failureErr))
		return
	}

	delay := w.calculateBackoff(retryCount)
	nextRetry := time.Now().Add(delay)

	w.logger.Info("scheduling retry", "id", msg.ID, "attempt", retryCount, "delay", delay)

	if err := w.queueRepo.UpdateStatus(ctx, msg.ID, domain.QueueStatusRetrying, retryCount, nextRetry, failureErr.Error()); err != nil {
		w.logger.Error("failed to update message retry status", "id", msg.ID, "error", err)
	}
	w.metrics.IncrementOutboundFailedTransient()
}

func (w *DeliveryWorker) handlePermanentFailure(ctx context.Context, msg *domain.OutboundMessage, reason string) {
	w.logger.Error("permanent delivery failure", "id", msg.ID, "reason", reason)

	if err := w.queueRepo.UpdateStatus(ctx, msg.ID, domain.QueueStatusFailed, msg.RetryCount, time.Time{}, reason); err != nil {
		w.logger.Error("failed to mark message as failed", "id", msg.ID, "error", err)
	}
	w.metrics.IncrementOutboundFailedPermanent()
}

func (w *DeliveryWorker) calculateBackoff(attempt int) time.Duration {
	// Exponential backoff: 1m * 2^(n-1), capped or specific steps
	switch attempt {
	case 1:
		return 1 * time.Minute
	case 2:
		return 5 * time.Minute
	case 3:
		return 15 * time.Minute
	case 4:
		return 1 * time.Hour
	case 5:
		return 6 * time.Hour
	default:
		// Exponential for further attempts, capped at 24h
		pow := float64(attempt - 5)
		delay := 6 * time.Hour * time.Duration(math.Pow(2, pow))
		if delay > 24*time.Hour {
			return 24 * time.Hour
		}
		return delay
	}
}
