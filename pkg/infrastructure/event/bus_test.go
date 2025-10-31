package event

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	logpkg "github.com/sage-x-project/blockchain-indexer/pkg/infrastructure/logger"
	"github.com/sage-x-project/blockchain-indexer/pkg/domain/models"
)

func TestNewEventBus(t *testing.T) {
	bus := NewEventBus(nil, nil)
	if bus == nil {
		t.Fatal("NewEventBus returned nil")
	}
}

func TestEventBus_Lifecycle(t *testing.T) {
	bus := NewEventBus(nil, nil)

	// Start
	if err := bus.Start(); err != nil {
		t.Fatalf("Failed to start event bus: %v", err)
	}

	// Start again should fail
	if err := bus.Start(); err == nil {
		t.Error("Expected error when starting already started bus")
	}

	// Stop
	if err := bus.Stop(); err != nil {
		t.Fatalf("Failed to stop event bus: %v", err)
	}
}

func TestEventBus_PublishSync(t *testing.T) {
	config := &EventBusConfig{
		QueueSize:           100,
		SubscriberQueueSize: 100,
		WorkerCount:         2,
		EnableMetrics:       true,
		MaxDeliveryLatency:  1 * time.Second,
	}

	bus := NewEventBus(config, nil)
	if err := bus.Start(); err != nil {
		t.Fatalf("Failed to start: %v", err)
	}
	defer bus.Stop()

	event := NewEvent(EventTypeBlockIndexed, "ethereum", &BlockIndexedPayload{
		Block: &models.Block{
			ChainID: "ethereum",
			Number:  1000,
		},
		TransactionCount: 10,
		ProcessingTime:   100 * time.Millisecond,
	})

	if err := bus.Publish(event); err != nil {
		t.Errorf("Failed to publish event: %v", err)
	}

	// Wait a bit for processing
	time.Sleep(100 * time.Millisecond)

	stats := bus.Stats()
	if stats.EventsPublished == 0 {
		t.Error("Event was not recorded as published")
	}
}

func TestEventBus_PublishAsync(t *testing.T) {
	config := &EventBusConfig{
		QueueSize:           100,
		SubscriberQueueSize: 100,
		WorkerCount:         2,
		EnableMetrics:       true,
		MaxDeliveryLatency:  1 * time.Second,
	}

	bus := NewEventBus(config, nil)
	if err := bus.Start(); err != nil {
		t.Fatalf("Failed to start: %v", err)
	}
	defer bus.Stop()

	event := NewEvent(EventTypeBlockIndexed, "ethereum", &BlockIndexedPayload{
		Block: &models.Block{
			ChainID: "ethereum",
			Number:  1000,
		},
		TransactionCount: 10,
		ProcessingTime:   100 * time.Millisecond,
	})

	bus.PublishAsync(event)

	// Wait a bit for processing
	time.Sleep(100 * time.Millisecond)

	stats := bus.Stats()
	if stats.EventsPublished == 0 {
		t.Error("Event was not recorded as published")
	}
}

func TestEventBus_SubscribeAll(t *testing.T) {
	config := &EventBusConfig{
		QueueSize:           100,
		SubscriberQueueSize: 100,
		WorkerCount:         2,
		EnableMetrics:       true,
		MaxDeliveryLatency:  1 * time.Second,
	}

	bus := NewEventBus(config, nil)
	if err := bus.Start(); err != nil {
		t.Fatalf("Failed to start: %v", err)
	}
	defer bus.Stop()

	var receivedCount atomic.Uint64
	var wg sync.WaitGroup
	wg.Add(1)

	handler := func(event *Event) {
		receivedCount.Add(1)
		if receivedCount.Load() == 1 {
			wg.Done()
		}
	}

	subID, err := bus.SubscribeAll(handler)
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	// Publish event
	event := NewEvent(EventTypeBlockIndexed, "ethereum", &BlockIndexedPayload{
		Block: &models.Block{
			ChainID: "ethereum",
			Number:  1000,
		},
	})

	bus.PublishAsync(event)

	// Wait for event to be received
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for event")
	}

	if receivedCount.Load() != 1 {
		t.Errorf("Expected 1 received event, got %d", receivedCount.Load())
	}

	// Unsubscribe
	if err := bus.Unsubscribe(subID); err != nil {
		t.Errorf("Failed to unsubscribe: %v", err)
	}
}

func TestEventBus_SubscribeType(t *testing.T) {
	config := &EventBusConfig{
		QueueSize:           100,
		SubscriberQueueSize: 100,
		WorkerCount:         2,
		EnableMetrics:       true,
		MaxDeliveryLatency:  1 * time.Second,
	}

	bus := NewEventBus(config, nil)
	if err := bus.Start(); err != nil {
		t.Fatalf("Failed to start: %v", err)
	}
	defer bus.Stop()

	var receivedCount atomic.Uint64
	var wg sync.WaitGroup
	wg.Add(1)

	handler := func(event *Event) {
		if event.Type != EventTypeBlockIndexed {
			t.Errorf("Expected EventTypeBlockIndexed, got %v", event.Type)
		}
		receivedCount.Add(1)
		if receivedCount.Load() == 1 {
			wg.Done()
		}
	}

	_, err := bus.SubscribeType(EventTypeBlockIndexed, handler)
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	// Publish matching event
	event1 := NewEvent(EventTypeBlockIndexed, "ethereum", &BlockIndexedPayload{
		Block: &models.Block{ChainID: "ethereum", Number: 1000},
	})
	bus.PublishAsync(event1)

	// Publish non-matching event
	event2 := NewEvent(EventTypeTransactionIndexed, "ethereum", &TransactionIndexedPayload{
		Transaction: &models.Transaction{ChainID: "ethereum", Hash: "0x123"},
		BlockNumber: 1000,
	})
	bus.PublishAsync(event2)

	// Wait for event to be received
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for event")
	}

	// Give some time for the second event (should not be received)
	time.Sleep(200 * time.Millisecond)

	if receivedCount.Load() != 1 {
		t.Errorf("Expected 1 received event, got %d", receivedCount.Load())
	}
}

func TestEventBus_SubscribeChain(t *testing.T) {
	config := &EventBusConfig{
		QueueSize:           100,
		SubscriberQueueSize: 100,
		WorkerCount:         2,
		EnableMetrics:       true,
		MaxDeliveryLatency:  1 * time.Second,
	}

	bus := NewEventBus(config, nil)
	if err := bus.Start(); err != nil {
		t.Fatalf("Failed to start: %v", err)
	}
	defer bus.Stop()

	var receivedCount atomic.Uint64
	var wg sync.WaitGroup
	wg.Add(1)

	handler := func(event *Event) {
		if event.ChainID != "ethereum" {
			t.Errorf("Expected ethereum, got %v", event.ChainID)
		}
		receivedCount.Add(1)
		if receivedCount.Load() == 1 {
			wg.Done()
		}
	}

	_, err := bus.SubscribeChain("ethereum", handler)
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	// Publish matching event
	event1 := NewEvent(EventTypeBlockIndexed, "ethereum", &BlockIndexedPayload{
		Block: &models.Block{ChainID: "ethereum", Number: 1000},
	})
	bus.PublishAsync(event1)

	// Publish non-matching event
	event2 := NewEvent(EventTypeBlockIndexed, "solana", &BlockIndexedPayload{
		Block: &models.Block{ChainID: "solana", Number: 1000},
	})
	bus.PublishAsync(event2)

	// Wait for event to be received
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for event")
	}

	// Give some time for the second event (should not be received)
	time.Sleep(200 * time.Millisecond)

	if receivedCount.Load() != 1 {
		t.Errorf("Expected 1 received event, got %d", receivedCount.Load())
	}
}

func TestEventBus_CustomFilter(t *testing.T) {
	config := &EventBusConfig{
		QueueSize:           100,
		SubscriberQueueSize: 100,
		WorkerCount:         2,
		EnableMetrics:       true,
		MaxDeliveryLatency:  1 * time.Second,
	}

	bus := NewEventBus(config, nil)
	if err := bus.Start(); err != nil {
		t.Fatalf("Failed to start: %v", err)
	}
	defer bus.Stop()

	var receivedCount atomic.Uint64
	var wg sync.WaitGroup
	wg.Add(1)

	handler := func(event *Event) {
		receivedCount.Add(1)
		if receivedCount.Load() == 1 {
			wg.Done()
		}
	}

	// Subscribe with custom filter (only block numbers > 1000)
	filter := func(event *Event) bool {
		if event.Type != EventTypeBlockIndexed {
			return false
		}
		if payload, ok := event.Payload.(*BlockIndexedPayload); ok {
			return payload.Block.Number > 1000
		}
		return false
	}

	_, err := bus.Subscribe(filter, handler)
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	// Publish non-matching event (block 900)
	event1 := NewEvent(EventTypeBlockIndexed, "ethereum", &BlockIndexedPayload{
		Block: &models.Block{ChainID: "ethereum", Number: 900},
	})
	bus.PublishAsync(event1)

	// Publish matching event (block 2000)
	event2 := NewEvent(EventTypeBlockIndexed, "ethereum", &BlockIndexedPayload{
		Block: &models.Block{ChainID: "ethereum", Number: 2000},
	})
	bus.PublishAsync(event2)

	// Wait for event to be received
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for event")
	}

	// Give some time for the first event (should not be received)
	time.Sleep(200 * time.Millisecond)

	if receivedCount.Load() != 1 {
		t.Errorf("Expected 1 received event, got %d", receivedCount.Load())
	}
}

func TestEventBus_MultipleSubscribers(t *testing.T) {
	config := &EventBusConfig{
		QueueSize:           100,
		SubscriberQueueSize: 100,
		WorkerCount:         2,
		EnableMetrics:       true,
		MaxDeliveryLatency:  1 * time.Second,
	}

	bus := NewEventBus(config, nil)
	if err := bus.Start(); err != nil {
		t.Fatalf("Failed to start: %v", err)
	}
	defer bus.Stop()

	var wg sync.WaitGroup
	wg.Add(3)

	received1 := atomic.Uint64{}
	received2 := atomic.Uint64{}
	received3 := atomic.Uint64{}

	handler1 := func(event *Event) {
		received1.Add(1)
		if received1.Load() == 1 {
			wg.Done()
		}
	}

	handler2 := func(event *Event) {
		received2.Add(1)
		if received2.Load() == 1 {
			wg.Done()
		}
	}

	handler3 := func(event *Event) {
		received3.Add(1)
		if received3.Load() == 1 {
			wg.Done()
		}
	}

	// Subscribe all three handlers
	if _, err := bus.SubscribeAll(handler1); err != nil {
		t.Fatalf("Failed to subscribe handler1: %v", err)
	}
	if _, err := bus.SubscribeAll(handler2); err != nil {
		t.Fatalf("Failed to subscribe handler2: %v", err)
	}
	if _, err := bus.SubscribeAll(handler3); err != nil {
		t.Fatalf("Failed to subscribe handler3: %v", err)
	}

	// Publish one event
	event := NewEvent(EventTypeBlockIndexed, "ethereum", &BlockIndexedPayload{
		Block: &models.Block{ChainID: "ethereum", Number: 1000},
	})
	bus.PublishAsync(event)

	// Wait for all handlers to receive the event
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for events")
	}

	// Verify all received
	if received1.Load() != 1 || received2.Load() != 1 || received3.Load() != 1 {
		t.Errorf("Expected all handlers to receive 1 event, got %d, %d, %d",
			received1.Load(), received2.Load(), received3.Load())
	}
}

func TestEventBus_Unsubscribe(t *testing.T) {
	config := &EventBusConfig{
		QueueSize:           100,
		SubscriberQueueSize: 100,
		WorkerCount:         2,
		EnableMetrics:       true,
		MaxDeliveryLatency:  1 * time.Second,
	}

	bus := NewEventBus(config, nil)
	if err := bus.Start(); err != nil {
		t.Fatalf("Failed to start: %v", err)
	}
	defer bus.Stop()

	var receivedCount atomic.Uint64

	handler := func(event *Event) {
		receivedCount.Add(1)
	}

	subID, err := bus.SubscribeAll(handler)
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	// Publish event
	event1 := NewEvent(EventTypeBlockIndexed, "ethereum", &BlockIndexedPayload{
		Block: &models.Block{ChainID: "ethereum", Number: 1000},
	})
	bus.PublishAsync(event1)

	// Wait a bit
	time.Sleep(200 * time.Millisecond)

	// Unsubscribe
	if err := bus.Unsubscribe(subID); err != nil {
		t.Errorf("Failed to unsubscribe: %v", err)
	}

	count1 := receivedCount.Load()

	// Publish another event
	event2 := NewEvent(EventTypeBlockIndexed, "ethereum", &BlockIndexedPayload{
		Block: &models.Block{ChainID: "ethereum", Number: 2000},
	})
	bus.PublishAsync(event2)

	// Wait a bit
	time.Sleep(200 * time.Millisecond)

	count2 := receivedCount.Load()

	// Count should not have increased
	if count2 != count1 {
		t.Errorf("Event received after unsubscribe: before=%d, after=%d", count1, count2)
	}
}

func TestEventBus_HandlerPanic(t *testing.T) {
	config := &EventBusConfig{
		QueueSize:           100,
		SubscriberQueueSize: 100,
		WorkerCount:         2,
		EnableMetrics:       true,
		MaxDeliveryLatency:  1 * time.Second,
	}

	logger, _ := logpkg.New(&logpkg.Config{
		Level:  "error",
		Format: "json",
	})

	bus := NewEventBus(config, logger)
	if err := bus.Start(); err != nil {
		t.Fatalf("Failed to start: %v", err)
	}
	defer bus.Stop()

	var wg sync.WaitGroup
	wg.Add(1)

	panicHandler := func(event *Event) {
		panic("intentional panic")
	}

	normalHandler := func(event *Event) {
		wg.Done()
	}

	// Subscribe both handlers
	if _, err := bus.SubscribeAll(panicHandler); err != nil {
		t.Fatalf("Failed to subscribe panic handler: %v", err)
	}
	if _, err := bus.SubscribeAll(normalHandler); err != nil {
		t.Fatalf("Failed to subscribe normal handler: %v", err)
	}

	// Publish event
	event := NewEvent(EventTypeBlockIndexed, "ethereum", &BlockIndexedPayload{
		Block: &models.Block{ChainID: "ethereum", Number: 1000},
	})
	bus.PublishAsync(event)

	// Normal handler should still receive event despite panic
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success - event bus didn't crash
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for event - event bus may have crashed")
	}
}

func TestEventBus_Stats(t *testing.T) {
	config := &EventBusConfig{
		QueueSize:           100,
		SubscriberQueueSize: 100,
		WorkerCount:         2,
		EnableMetrics:       true,
		MaxDeliveryLatency:  1 * time.Second,
	}

	bus := NewEventBus(config, nil)
	if err := bus.Start(); err != nil {
		t.Fatalf("Failed to start: %v", err)
	}
	defer bus.Stop()

	var wg sync.WaitGroup
	wg.Add(1)

	handler := func(event *Event) {
		wg.Done()
	}

	_, err := bus.SubscribeAll(handler)
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	// Publish event
	event := NewEvent(EventTypeBlockIndexed, "ethereum", &BlockIndexedPayload{
		Block: &models.Block{ChainID: "ethereum", Number: 1000},
	})
	bus.PublishAsync(event)

	// Wait for delivery
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for event")
	}

	stats := bus.Stats()

	if stats.EventsPublished != 1 {
		t.Errorf("Expected 1 published event, got %d", stats.EventsPublished)
	}

	if stats.EventsDelivered != 1 {
		t.Errorf("Expected 1 delivered event, got %d", stats.EventsDelivered)
	}

	if stats.ActiveSubscribers != 1 {
		t.Errorf("Expected 1 active subscriber, got %d", stats.ActiveSubscribers)
	}
}

func TestEventBus_Concurrent(t *testing.T) {
	config := &EventBusConfig{
		QueueSize:           1000,
		SubscriberQueueSize: 1000,
		WorkerCount:         10,
		EnableMetrics:       true,
		MaxDeliveryLatency:  5 * time.Second,
	}

	bus := NewEventBus(config, nil)
	if err := bus.Start(); err != nil {
		t.Fatalf("Failed to start: %v", err)
	}
	defer bus.Stop()

	const numPublishers = 10
	const numSubscribers = 5
	const eventsPerPublisher = 100

	var receivedCount atomic.Uint64
	var wg sync.WaitGroup

	// Start subscribers
	for i := 0; i < numSubscribers; i++ {
		handler := func(event *Event) {
			receivedCount.Add(1)
		}
		if _, err := bus.SubscribeAll(handler); err != nil {
			t.Fatalf("Failed to subscribe: %v", err)
		}
	}

	// Start publishers
	wg.Add(numPublishers)
	for i := 0; i < numPublishers; i++ {
		go func(publisherID int) {
			defer wg.Done()
			for j := 0; j < eventsPerPublisher; j++ {
				event := NewEvent(EventTypeBlockIndexed, "ethereum", &BlockIndexedPayload{
					Block: &models.Block{
						ChainID: "ethereum",
						Number:  uint64(publisherID*1000 + j),
					},
				})
				bus.PublishAsync(event)
			}
		}(i)
	}

	// Wait for publishers to finish
	wg.Wait()

	// Wait for all events to be processed
	time.Sleep(2 * time.Second)

	// Each event should be delivered to each subscriber
	expectedTotal := uint64(numPublishers * eventsPerPublisher * numSubscribers)
	actualTotal := receivedCount.Load()

	// Allow some tolerance for timing
	if actualTotal < expectedTotal*90/100 {
		t.Errorf("Expected approximately %d events, got %d", expectedTotal, actualTotal)
	}

	stats := bus.Stats()
	t.Logf("Stats: Published=%d, Delivered=%d, Dropped=%d",
		stats.EventsPublished, stats.EventsDelivered, stats.EventsDropped)
}
