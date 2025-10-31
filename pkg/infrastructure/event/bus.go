package event

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	logpkg "github.com/sage-x-project/blockchain-indexer/pkg/infrastructure/logger"
	"go.uber.org/zap"
)

// subscription represents an active event subscription
type subscription struct {
	id      SubscriptionID
	filter  EventFilter
	handler EventHandler
	queue   chan *Event
	ctx     context.Context
	cancel  context.CancelFunc
}

// eventBus is the main implementation of EventBus
type eventBus struct {
	config *EventBusConfig

	// Subscription management
	mu            sync.RWMutex
	subscriptions map[SubscriptionID]*subscription
	nextID        uint64

	// Event delivery
	publishQueue chan *Event

	// Lifecycle
	started atomic.Bool
	stopOnce sync.Once
	stopChan chan struct{}
	wg       sync.WaitGroup

	// Statistics
	stats eventBusStats

	// Logger
	logger *logpkg.Logger
}

// eventBusStats holds atomic counters for statistics
type eventBusStats struct {
	eventsPublished   atomic.Uint64
	eventsDelivered   atomic.Uint64
	eventsDropped     atomic.Uint64
	totalLatencyNs    atomic.Uint64
	latencySamples    atomic.Uint64
	activeSubscribers atomic.Int32
}

// EventBusConfig holds event bus configuration
type EventBusConfig struct {
	// QueueSize is the size of the main publish queue
	QueueSize int

	// SubscriberQueueSize is the size of each subscriber's queue
	SubscriberQueueSize int

	// WorkerCount is the number of workers processing the publish queue
	WorkerCount int

	// EnableMetrics enables detailed metrics collection
	EnableMetrics bool

	// MaxDeliveryLatency is the maximum time to wait for event delivery
	MaxDeliveryLatency time.Duration
}

// DefaultEventBusConfig returns default configuration
func DefaultEventBusConfig() *EventBusConfig {
	return &EventBusConfig{
		QueueSize:           10000,
		SubscriberQueueSize: 1000,
		WorkerCount:         10,
		EnableMetrics:       true,
		MaxDeliveryLatency:  5 * time.Second,
	}
}

// NewEventBus creates a new event bus
func NewEventBus(config *EventBusConfig, logger *logpkg.Logger) EventBus {
	if config == nil {
		config = DefaultEventBusConfig()
	}

	if logger == nil {
		// Create default logger
		defaultLogger, err := logpkg.New(&logpkg.Config{
			Level:  "info",
			Format: "json",
		})
		if err != nil {
			panic(fmt.Sprintf("failed to create default logger: %v", err))
		}
		logger = defaultLogger
	}

	bus := &eventBus{
		config:        config,
		subscriptions: make(map[SubscriptionID]*subscription),
		publishQueue:  make(chan *Event, config.QueueSize),
		stopChan:      make(chan struct{}),
		logger:        logger,
	}

	return bus
}

// Start starts the event bus
func (b *eventBus) Start() error {
	if !b.started.CompareAndSwap(false, true) {
		return fmt.Errorf("event bus already started")
	}

	b.logger.Info("starting event bus",
		zap.Int("queue_size", b.config.QueueSize),
		zap.Int("workers", b.config.WorkerCount),
	)

	// Start worker goroutines
	for i := 0; i < b.config.WorkerCount; i++ {
		b.wg.Add(1)
		go b.worker(i)
	}

	b.logger.Info("event bus started")
	return nil
}

// Stop stops the event bus
func (b *eventBus) Stop() error {
	if !b.started.Load() {
		return fmt.Errorf("event bus not started")
	}

	b.stopOnce.Do(func() {
		b.logger.Info("stopping event bus")

		// Cancel all subscriptions first
		b.mu.Lock()
		for _, sub := range b.subscriptions {
			sub.cancel()
			close(sub.queue)
		}
		b.subscriptions = make(map[SubscriptionID]*subscription)
		b.mu.Unlock()

		// Signal stop to workers
		close(b.stopChan)

		// Wait for all goroutines to finish
		b.wg.Wait()

		b.started.Store(false)
		b.logger.Info("event bus stopped")
	})

	return nil
}

// Publish publishes an event synchronously
func (b *eventBus) Publish(event *Event) error {
	if !b.started.Load() {
		return fmt.Errorf("event bus not started")
	}

	if event == nil {
		return fmt.Errorf("event cannot be nil")
	}

	select {
	case b.publishQueue <- event:
		b.stats.eventsPublished.Add(1)
		return nil
	case <-time.After(b.config.MaxDeliveryLatency):
		b.stats.eventsDropped.Add(1)
		return fmt.Errorf("failed to publish event: queue full")
	}
}

// PublishAsync publishes an event asynchronously
func (b *eventBus) PublishAsync(event *Event) {
	if !b.started.Load() {
		b.logger.Warn("cannot publish to stopped event bus")
		return
	}

	if event == nil {
		b.logger.Warn("attempted to publish nil event")
		return
	}

	select {
	case b.publishQueue <- event:
		b.stats.eventsPublished.Add(1)
	default:
		// Drop event if queue is full
		b.stats.eventsDropped.Add(1)
		b.logger.Warn("event dropped: publish queue full",
			zap.String("event_type", event.Type.String()),
			zap.String("event_id", event.ID),
		)
	}
}

// Subscribe subscribes to events matching the filter
func (b *eventBus) Subscribe(filter EventFilter, handler EventHandler) (SubscriptionID, error) {
	if !b.started.Load() {
		return "", fmt.Errorf("event bus not started")
	}

	if filter == nil {
		return "", fmt.Errorf("filter cannot be nil")
	}

	if handler == nil {
		return "", fmt.Errorf("handler cannot be nil")
	}

	return b.addSubscription(filter, handler)
}

// SubscribeType subscribes to events of a specific type
func (b *eventBus) SubscribeType(eventType EventType, handler EventHandler) (SubscriptionID, error) {
	filter := func(event *Event) bool {
		return event.Type == eventType
	}
	return b.Subscribe(filter, handler)
}

// SubscribeChain subscribes to events for a specific chain
func (b *eventBus) SubscribeChain(chainID string, handler EventHandler) (SubscriptionID, error) {
	filter := func(event *Event) bool {
		return event.ChainID == chainID
	}
	return b.Subscribe(filter, handler)
}

// SubscribeAll subscribes to all events
func (b *eventBus) SubscribeAll(handler EventHandler) (SubscriptionID, error) {
	filter := func(event *Event) bool {
		return true
	}
	return b.Subscribe(filter, handler)
}

// Unsubscribe unsubscribes from events
func (b *eventBus) Unsubscribe(id SubscriptionID) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	sub, exists := b.subscriptions[id]
	if !exists {
		return fmt.Errorf("subscription not found: %s", id)
	}

	// Cancel subscription context
	sub.cancel()

	// Close subscription queue
	close(sub.queue)

	// Remove from map
	delete(b.subscriptions, id)

	b.stats.activeSubscribers.Add(-1)

	b.logger.Debug("subscription removed", zap.String("subscription_id", string(id)))

	return nil
}

// Stats returns event bus statistics
func (b *eventBus) Stats() *EventBusStats {
	var avgLatency time.Duration
	samples := b.stats.latencySamples.Load()
	if samples > 0 {
		avgNs := b.stats.totalLatencyNs.Load() / samples
		avgLatency = time.Duration(avgNs)
	}

	b.mu.RLock()
	queueLen := len(b.publishQueue)
	b.mu.RUnlock()

	return &EventBusStats{
		EventsPublished:   b.stats.eventsPublished.Load(),
		EventsDelivered:   b.stats.eventsDelivered.Load(),
		EventsDropped:     b.stats.eventsDropped.Load(),
		ActiveSubscribers: int(b.stats.activeSubscribers.Load()),
		AverageLatency:    avgLatency,
		QueueLength:       queueLen,
		QueueCapacity:     b.config.QueueSize,
	}
}

// worker processes events from the publish queue
func (b *eventBus) worker(id int) {
	defer b.wg.Done()

	b.logger.Debug("event bus worker started", zap.Int("worker_id", id))

	for {
		select {
		case <-b.stopChan:
			b.logger.Debug("event bus worker stopped", zap.Int("worker_id", id))
			return

		case event := <-b.publishQueue:
			if event == nil {
				continue
			}

			startTime := time.Now()
			b.deliverEvent(event)

			// Record latency
			if b.config.EnableMetrics {
				latency := time.Since(startTime)
				b.stats.totalLatencyNs.Add(uint64(latency.Nanoseconds()))
				b.stats.latencySamples.Add(1)
			}
		}
	}
}

// deliverEvent delivers an event to all matching subscribers
func (b *eventBus) deliverEvent(event *Event) {
	b.mu.RLock()
	subscribers := make([]*subscription, 0, len(b.subscriptions))
	for _, sub := range b.subscriptions {
		subscribers = append(subscribers, sub)
	}
	b.mu.RUnlock()

	for _, sub := range subscribers {
		// Check if subscription is still active
		select {
		case <-sub.ctx.Done():
			continue
		default:
		}

		// Apply filter
		if sub.filter != nil && !sub.filter(event) {
			continue
		}

		// Deliver event to subscriber
		select {
		case sub.queue <- event:
			b.stats.eventsDelivered.Add(1)
		default:
			// Drop event if subscriber queue is full
			b.stats.eventsDropped.Add(1)
			b.logger.Warn("event dropped: subscriber queue full",
				zap.String("subscription_id", string(sub.id)),
				zap.String("event_type", event.Type.String()),
				zap.String("event_id", event.ID),
			)
		}
	}
}

// addSubscription adds a new subscription
func (b *eventBus) addSubscription(filter EventFilter, handler EventHandler) (SubscriptionID, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Generate subscription ID
	id := SubscriptionID(fmt.Sprintf("sub-%d", b.nextID))
	b.nextID++

	// Create subscription context
	ctx, cancel := context.WithCancel(context.Background())

	// Create subscription
	sub := &subscription{
		id:      id,
		filter:  filter,
		handler: handler,
		queue:   make(chan *Event, b.config.SubscriberQueueSize),
		ctx:     ctx,
		cancel:  cancel,
	}

	// Add to map
	b.subscriptions[id] = sub

	b.stats.activeSubscribers.Add(1)

	// Start handler goroutine
	b.wg.Add(1)
	go b.handleSubscription(sub)

	b.logger.Debug("subscription added", zap.String("subscription_id", string(id)))

	return id, nil
}

// handleSubscription processes events for a subscription
func (b *eventBus) handleSubscription(sub *subscription) {
	defer b.wg.Done()

	for {
		select {
		case <-sub.ctx.Done():
			return

		case event, ok := <-sub.queue:
			if !ok {
				return
			}

			// Call handler
			// Recover from panics in handler
			func() {
				defer func() {
					if r := recover(); r != nil {
						b.logger.Error("handler panicked",
							zap.String("subscription_id", string(sub.id)),
							zap.String("event_type", event.Type.String()),
							zap.String("event_id", event.ID),
							zap.Any("panic", r),
						)
					}
				}()

				sub.handler(event)
			}()
		}
	}
}
