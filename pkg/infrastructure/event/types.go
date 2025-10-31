package event

import (
	"time"

	"github.com/sage-x-project/blockchain-indexer/pkg/domain/models"
)

// EventType represents the type of event
type EventType string

const (
	// EventTypeBlockIndexed is emitted when a block is indexed
	EventTypeBlockIndexed EventType = "block.indexed"

	// EventTypeBlockProcessed is emitted when a block is processed
	EventTypeBlockProcessed EventType = "block.processed"

	// EventTypeTransactionIndexed is emitted when a transaction is indexed
	EventTypeTransactionIndexed EventType = "transaction.indexed"

	// EventTypeTransactionProcessed is emitted when a transaction is processed
	EventTypeTransactionProcessed EventType = "transaction.processed"

	// EventTypeChainSyncStarted is emitted when chain sync starts
	EventTypeChainSyncStarted EventType = "chain.sync.started"

	// EventTypeChainSyncCompleted is emitted when chain sync completes
	EventTypeChainSyncCompleted EventType = "chain.sync.completed"

	// EventTypeChainSyncError is emitted when chain sync encounters an error
	EventTypeChainSyncError EventType = "chain.sync.error"

	// EventTypeGapDetected is emitted when a gap is detected
	EventTypeGapDetected EventType = "gap.detected"

	// EventTypeGapRecovered is emitted when a gap is recovered
	EventTypeGapRecovered EventType = "gap.recovered"
)

// String returns the string representation of EventType
func (e EventType) String() string {
	return string(e)
}

// Event represents a system event
type Event struct {
	ID        string                 // Unique event ID
	Type      EventType              // Event type
	Timestamp time.Time              // Event timestamp
	ChainID   string                 // Chain ID
	Payload   interface{}            // Event payload
	Metadata  map[string]interface{} // Additional metadata
}

// NewEvent creates a new event
func NewEvent(eventType EventType, chainID string, payload interface{}) *Event {
	return &Event{
		ID:        generateEventID(),
		Type:      eventType,
		Timestamp: time.Now(),
		ChainID:   chainID,
		Payload:   payload,
		Metadata:  make(map[string]interface{}),
	}
}

// WithMetadata adds metadata to the event
func (e *Event) WithMetadata(key string, value interface{}) *Event {
	if e.Metadata == nil {
		e.Metadata = make(map[string]interface{})
	}
	e.Metadata[key] = value
	return e
}

// BlockIndexedPayload is the payload for block indexed events
type BlockIndexedPayload struct {
	Block            *models.Block
	TransactionCount int
	ProcessingTime   time.Duration
}

// TransactionIndexedPayload is the payload for transaction indexed events
type TransactionIndexedPayload struct {
	Transaction *models.Transaction
	BlockNumber uint64
}

// ChainSyncPayload is the payload for chain sync events
type ChainSyncPayload struct {
	ChainID            string
	StartBlock         uint64
	EndBlock           uint64
	CurrentBlock       uint64
	TotalBlocks        uint64
	ProgressPercentage float64
}

// GapPayload is the payload for gap events
type GapPayload struct {
	ChainID    string
	StartBlock uint64
	EndBlock   uint64
	Size       uint64
}

// ErrorPayload is the payload for error events
type ErrorPayload struct {
	Error   error
	Context string
	Details map[string]interface{}
}

// EventHandler is a function that handles events
type EventHandler func(*Event)

// EventFilter is a function that filters events
type EventFilter func(*Event) bool

// Publisher publishes events
type Publisher interface {
	// Publish publishes an event
	Publish(event *Event) error

	// PublishAsync publishes an event asynchronously
	PublishAsync(event *Event)
}

// Subscriber subscribes to events
type Subscriber interface {
	// Subscribe subscribes to events matching the filter
	Subscribe(filter EventFilter, handler EventHandler) (SubscriptionID, error)

	// SubscribeType subscribes to events of a specific type
	SubscribeType(eventType EventType, handler EventHandler) (SubscriptionID, error)

	// SubscribeChain subscribes to events for a specific chain
	SubscribeChain(chainID string, handler EventHandler) (SubscriptionID, error)

	// SubscribeAll subscribes to all events
	SubscribeAll(handler EventHandler) (SubscriptionID, error)

	// Unsubscribe unsubscribes from events
	Unsubscribe(id SubscriptionID) error
}

// EventBus combines Publisher and Subscriber
type EventBus interface {
	Publisher
	Subscriber

	// Start starts the event bus
	Start() error

	// Stop stops the event bus
	Stop() error

	// Stats returns event bus statistics
	Stats() *EventBusStats
}

// SubscriptionID is a unique subscription identifier
type SubscriptionID string

// EventBusStats represents event bus statistics
type EventBusStats struct {
	EventsPublished   uint64
	EventsDelivered   uint64
	EventsDropped     uint64
	ActiveSubscribers int
	AverageLatency    time.Duration
	QueueLength       int
	QueueCapacity     int
}

// generateEventID generates a unique event ID
func generateEventID() string {
	// Simple timestamp-based ID
	// In production, you might want to use UUID or similar
	return time.Now().Format("20060102150405.000000")
}
