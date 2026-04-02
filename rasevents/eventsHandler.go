package rasevents

//goland:noinspection GoSnakeCaseUsage
import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	nats_service "github.com/transactrx/nats-service/pkg/nats-service"
	nats_service_client "github.com/transactrx/nats-service/pkg/nats-service-client"
)

// Config holds the configuration for the events handler.
type Config struct {
	// DefaultNamespace is the default event namespace if none is provided.
	// Can be set via EVENTS_DEFAULT_NAMESPACE env var. Defaults to "PatientNotification".
	DefaultNamespace string

	// Subject is the base NATS subject for event collection.
	// Can be set via EVENTS_SUBJECT env var. Defaults to "trx.eventscollector.collect".
	Subject string

	// Timeout is the timeout for NATS requests.
	// Can be set via EVENTS_TIMEOUT_SECONDS env var. Defaults to 60 seconds.
	Timeout time.Duration

	// WorkerPoolSize is the number of workers for async event sending.
	// Can be set via EVENTS_WORKER_POOL_SIZE env var. Defaults to 50.
	WorkerPoolSize int

	// EventQueueSize is the size of the async event queue.
	// Can be set via EVENTS_QUEUE_SIZE env var. Defaults to 1000.
	EventQueueSize int
}

// DefaultConfig returns a Config with default values, reading from environment variables if set.
func DefaultConfig() Config {
	cfg := Config{
		DefaultNamespace: "PatientNotification",
		Subject:          "trx.eventscollector.collect",
		Timeout:          60 * time.Second,
		WorkerPoolSize:   50,
		EventQueueSize:   1000,
	}

	if ns := os.Getenv("EVENTS_DEFAULT_NAMESPACE"); ns != "" {
		cfg.DefaultNamespace = ns
	}
	if subj := os.Getenv("EVENTS_SUBJECT"); subj != "" {
		cfg.Subject = subj
	}
	if timeoutStr := os.Getenv("EVENTS_TIMEOUT_SECONDS"); timeoutStr != "" {
		if timeout, err := strconv.Atoi(timeoutStr); err == nil && timeout > 0 {
			cfg.Timeout = time.Duration(timeout) * time.Second
		}
	}
	if poolSizeStr := os.Getenv("EVENTS_WORKER_POOL_SIZE"); poolSizeStr != "" {
		if poolSize, err := strconv.Atoi(poolSizeStr); err == nil && poolSize > 0 {
			cfg.WorkerPoolSize = poolSize
		}
	}
	if queueSizeStr := os.Getenv("EVENTS_QUEUE_SIZE"); queueSizeStr != "" {
		if queueSize, err := strconv.Atoi(queueSizeStr); err == nil && queueSize > 0 {
			cfg.EventQueueSize = queueSize
		}
	}

	return cfg
}

// Event represents an event to be sent.
type Event struct {
	EventID      string      `json:"eventId"`
	EventTime    string      `json:"eventTime"`
	EventPayload interface{} `json:"eventPayload"`
}

// NatsClient defines the interface for NATS operations, enabling dependency injection.
type NatsClient interface {
	DoRequest(correlationId string, subject string, headers nats_service_client.Header, data []byte, timeout time.Duration) (*nats_service_client.NatsResponseMessage, *nats_service.NatsServiceError, error)
}

var (
	config     Config
	configOnce sync.Once

	natsClient   NatsClient
	natsClientMu sync.RWMutex
	natsInitErr  error

	// Worker pool for async event sending
	eventQueue    chan eventWork
	poolOnce      sync.Once
	poolStopCh    chan struct{}
	poolStoppedCh chan struct{}
	poolWg        sync.WaitGroup
)

type eventWork struct {
	ctx            context.Context
	eventNamespace string
	eventType      string
	payload        interface{}
}

// Init initializes the events handler with the given configuration.
// If cfg is nil, DefaultConfig() is used.
// This function is safe to call multiple times; subsequent calls are no-ops.
func Init(cfg *Config) {
	configOnce.Do(func() {
		if cfg != nil {
			config = *cfg
		} else {
			config = DefaultConfig()
		}
		slog.Info("Events handler initialized",
			"defaultNamespace", config.DefaultNamespace,
			"subject", config.Subject,
			"timeout", config.Timeout,
			"workerPoolSize", config.WorkerPoolSize,
			"eventQueueSize", config.EventQueueSize,
		)
	})
}

// ensureInit ensures the handler is initialized with defaults if Init hasn't been called.
func ensureInit() {
	Init(nil)
}

// GetNatsClient returns the NATS client, creating it if necessary.
// Returns an error if client creation fails. Safe to call concurrently.
func GetNatsClient() (NatsClient, error) {
	natsClientMu.RLock()
	if natsClient != nil {
		defer natsClientMu.RUnlock()
		return natsClient, nil
	}
	natsClientMu.RUnlock()

	natsClientMu.Lock()
	defer natsClientMu.Unlock()

	// Double-check after acquiring write lock
	if natsClient != nil {
		return natsClient, nil
	}

	client, err := nats_service_client.NewClient()
	if err != nil {
		natsInitErr = fmt.Errorf("error creating NATS client: %w", err)
		slog.Error("error creating NATS client", "error", err)
		return nil, natsInitErr
	}

	natsClient = client
	natsInitErr = nil
	return natsClient, nil
}

// SetNatsClient sets a custom NATS client (useful for testing).
func SetNatsClient(client NatsClient) {
	natsClientMu.Lock()
	defer natsClientMu.Unlock()
	natsClient = client
	natsInitErr = nil
}

// ResetNatsClient resets the NATS client, allowing reinitialization.
// Useful for testing or recovering from connection failures.
func ResetNatsClient() {
	natsClientMu.Lock()
	defer natsClientMu.Unlock()
	natsClient = nil
	natsInitErr = nil
}

// SendEvent sends an event synchronously and returns any error.
// Use SendEventWithContext for cancellation support.
func SendEvent(eventNamespace string, eventType string, payload interface{}) error {
	return SendEventWithContext(context.Background(), eventNamespace, eventType, payload)
}

// SendEventWithContext sends an event synchronously with context support.
func SendEventWithContext(ctx context.Context, eventNamespace string, eventType string, payload interface{}) error {
	ensureInit()

	client, err := GetNatsClient()
	if err != nil {
		return err
	}

	return sendEventWithClient(ctx, client, eventNamespace, eventType, payload)
}

// SendEventWithClient sends an event using the provided client.
// This is useful for dependency injection and testing.
func SendEventWithClient(ctx context.Context, client NatsClient, eventNamespace string, eventType string, payload interface{}) error {
	ensureInit()
	return sendEventWithClient(ctx, client, eventNamespace, eventType, payload)
}

// sendEventWithClient contains the core event sending logic.
func sendEventWithClient(ctx context.Context, client NatsClient, eventNamespace string, eventType string, payload interface{}) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	eventPayload, err := CreateEvent(payload)
	if err != nil {
		return fmt.Errorf("error creating event: %w", err)
	}

	if eventNamespace == "" {
		eventNamespace = config.DefaultNamespace
	}

	subject := config.Subject + "." + eventNamespace + "." + eventType
	_, nSvrErr, err := client.DoRequest("", subject, nil, eventPayload, config.Timeout)
	if err != nil {
		return fmt.Errorf("error sending event request: %w", err)
	}
	if nSvrErr != nil {
		return fmt.Errorf("NATS service error: status=%d, message=%s", nSvrErr.Status, nSvrErr.ErrorMessage)
	}

	return nil
}

// initWorkerPool initializes the worker pool for async event sending.
func initWorkerPool() {
	poolOnce.Do(func() {
		ensureInit()
		eventQueue = make(chan eventWork, config.EventQueueSize)
		poolStopCh = make(chan struct{})
		poolStoppedCh = make(chan struct{})

		poolWg.Add(config.WorkerPoolSize)
		for i := 0; i < config.WorkerPoolSize; i++ {
			go eventWorker()
		}

		// Goroutine to signal when all workers have stopped
		go func() {
			poolWg.Wait()
			close(poolStoppedCh)
		}()

		slog.Info("Event worker pool initialized", "workers", config.WorkerPoolSize, "queueSize", config.EventQueueSize)
	})
}

// eventWorker processes events from the queue.
func eventWorker() {
	defer poolWg.Done()
	for {
		select {
		case work, ok := <-eventQueue:
			if !ok {
				return // channel closed
			}
			if err := SendEventWithContext(work.ctx, work.eventNamespace, work.eventType, work.payload); err != nil {
				slog.Error("Error sending async event", "namespace", work.eventNamespace, "type", work.eventType, "error", err)
			}
		case <-poolStopCh:
			return
		}
	}
}

// SendEventAsync queues an event for asynchronous sending via the worker pool.
// This is the preferred method for fire-and-forget event sending.
// If the queue is full, the event is dropped with a warning log.
// Returns true if the event was queued, false if dropped.
func SendEventAsync(eventNamespace string, eventType string, payload interface{}) bool {
	return SendEventAsyncWithContext(context.Background(), eventNamespace, eventType, payload)
}

// SendEventAsyncWithContext queues an event for asynchronous sending with context support.
// Returns true if the event was queued, false if dropped.
func SendEventAsyncWithContext(ctx context.Context, eventNamespace string, eventType string, payload interface{}) bool {
	initWorkerPool()

	select {
	case eventQueue <- eventWork{ctx, eventNamespace, eventType, payload}:
		return true
	default:
		slog.Warn("Event queue full, dropping event", "namespace", eventNamespace, "type", eventType)
		return false
	}
}

// StopEventWorkerPool gracefully stops the event worker pool.
// Call during application shutdown.
func StopEventWorkerPool() {
	if poolStopCh != nil {
		close(poolStopCh)
		<-poolStoppedCh // wait for workers to finish
		slog.Info("Event worker pool stopped")
	}
}

// CreateEvent creates an event with the given payload.
func CreateEvent(payload interface{}) ([]byte, error) {
	event := Event{
		EventID:      uuid.New().String(),
		EventTime:    time.Now().Format(time.RFC3339),
		EventPayload: payload,
	}

	slog.Debug("CreateEvent", "event", event)

	bytes, err := json.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("error marshaling event: %w", err)
	}

	return bytes, nil
}

// GetConfig returns the current configuration.
// Returns a copy to prevent modification.
func GetConfig() Config {
	ensureInit()
	return config
}
