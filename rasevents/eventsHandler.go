// Package rasevents provides event publishing via NATS with synchronous and asynchronous support.
//
// Events can be sent synchronously using [SendEvent] or queued for asynchronous delivery
// using [SendEventAsync]. The package supports both a global handler for simple usage
// and independent [EventsHandler] instances for more control.
//
// Configuration can be provided via [Config] struct or environment variables
// (EVENTS_DEFAULT_NAMESPACE, EVENTS_SUBJECT, etc.).
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
	// DefaultNamespace is the default event namespace.
	// Can be set via EVENTS_DEFAULT_NAMESPACE env var.
	DefaultNamespace string

	// Subject is the base NATS subject for event collection.
	// Can be set via EVENTS_SUBJECT env var.
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

	// Hooks provides optional callbacks for observability.
	Hooks *Hooks
}

// Hooks provides callbacks for observability and metrics collection.
type Hooks struct {
	// OnEventSent is called after each synchronous event send attempt completes.
	// duration is the time taken to send the event.
	// err is nil on success.
	OnEventSent func(namespace, eventType string, duration time.Duration, err error)

	// OnEventQueued is called when an async event is submitted.
	// dropped is true if the queue was full and the event was discarded.
	OnEventQueued func(namespace, eventType string, dropped bool)

	// OnEventProcessed is called after an async worker processes an event.
	// err is nil on success.
	OnEventProcessed func(namespace, eventType string, duration time.Duration, err error)
}

// DefaultConfig returns a Config with default values, reading from environment variables if set.
// DefaultNamespace and Subject Required
func DefaultConfig() Config {
	cfg := Config{
		DefaultNamespace: "", // required
		Subject:          "", //required
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

// EventsHandler manages event publishing with its own configuration and client.
// Use NewEventsHandler to create an instance, or use the package-level functions
// for a shared global handler.
type EventsHandler struct {
	config Config

	clientMu  sync.RWMutex
	client    NatsClient
	clientErr error

	poolOnce     sync.Once
	eventQueue   chan eventWork
	poolStopCh   chan struct{}
	poolWg       sync.WaitGroup
	shutdownOnce sync.Once
	shutdownMu   sync.RWMutex
	shuttingDown bool
}

type eventWork struct {
	ctx            context.Context
	eventNamespace string
	eventType      string
	payload        interface{}
}

// NewEventsHandler creates a new EventsHandler with the given configuration.
// If client is nil, a client will be created lazily on first use.
func NewEventsHandler(cfg Config, client NatsClient) *EventsHandler {
	h := &EventsHandler{
		config: cfg,
		client: client,
	}
	slog.Info("Events handler created",
		"defaultNamespace", cfg.DefaultNamespace,
		"subject", cfg.Subject,
		"timeout", cfg.Timeout,
		"workerPoolSize", cfg.WorkerPoolSize,
		"eventQueueSize", cfg.EventQueueSize,
	)
	return h
}

// GetClient returns the NATS client, creating it if necessary.
func (h *EventsHandler) GetClient() (NatsClient, error) {
	h.clientMu.RLock()
	if h.client != nil {
		defer h.clientMu.RUnlock()
		return h.client, nil
	}
	h.clientMu.RUnlock()

	h.clientMu.Lock()
	defer h.clientMu.Unlock()

	if h.client != nil {
		return h.client, nil
	}

	client, err := nats_service_client.NewClient()
	if err != nil {
		h.clientErr = fmt.Errorf("error creating NATS client: %w", err)
		slog.Error("error creating NATS client", "error", err)
		return nil, h.clientErr
	}

	h.client = client
	h.clientErr = nil
	return h.client, nil
}

// SetClient sets a custom NATS client (useful for testing).
func (h *EventsHandler) SetClient(client NatsClient) {
	h.clientMu.Lock()
	defer h.clientMu.Unlock()
	h.client = client
	h.clientErr = nil
}

// GetConfig returns a copy of the current configuration.
func (h *EventsHandler) GetConfig() Config {
	return h.config
}

// SendEvent sends an event synchronously and returns any error.
func (h *EventsHandler) SendEvent(eventNamespace string, eventType string, payload interface{}) error {
	return h.SendEventWithContext(context.Background(), eventNamespace, eventType, payload)
}

// SendEventWithContext sends an event synchronously with context support.
func (h *EventsHandler) SendEventWithContext(ctx context.Context, eventNamespace string, eventType string, payload interface{}) error {
	client, err := h.GetClient()
	if err != nil {
		return err
	}
	return h.sendEventWithClient(ctx, client, eventNamespace, eventType, payload)
}

// SendEventWithClient sends an event using the provided client.
func (h *EventsHandler) SendEventWithClient(ctx context.Context, client NatsClient, eventNamespace string, eventType string, payload interface{}) error {
	return h.sendEventWithClient(ctx, client, eventNamespace, eventType, payload)
}

func (h *EventsHandler) sendEventWithClient(ctx context.Context, client NatsClient, eventNamespace string, eventType string, payload interface{}) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	start := time.Now()
	var sendErr error
	defer func() {
		if h.config.Hooks != nil && h.config.Hooks.OnEventSent != nil {
			h.config.Hooks.OnEventSent(eventNamespace, eventType, time.Since(start), sendErr)
		}
	}()

	eventPayload, err := CreateEvent(payload)
	if err != nil {
		sendErr = fmt.Errorf("error creating event: %w", err)
		return sendErr
	}

	if eventNamespace == "" {
		eventNamespace = h.config.DefaultNamespace
	}

	subject := h.config.Subject + "." + eventNamespace + "." + eventType
	_, nSvrErr, err := client.DoRequest("", subject, nil, eventPayload, h.config.Timeout)
	if err != nil {
		sendErr = fmt.Errorf("error sending event request: %w", err)
		return sendErr
	}
	if nSvrErr != nil {
		sendErr = fmt.Errorf("NATS service error: status=%d, message=%s", nSvrErr.Status, nSvrErr.ErrorMessage)
		return sendErr
	}

	return nil
}

// initWorkerPool initializes the worker pool for async event sending.
func (h *EventsHandler) initWorkerPool() {
	h.poolOnce.Do(func() {
		h.eventQueue = make(chan eventWork, h.config.EventQueueSize)
		h.poolStopCh = make(chan struct{})

		h.poolWg.Add(h.config.WorkerPoolSize)
		for i := 0; i < h.config.WorkerPoolSize; i++ {
			go h.eventWorker()
		}

		slog.Info("Event worker pool initialized", "workers", h.config.WorkerPoolSize, "queueSize", h.config.EventQueueSize)
	})
}

func (h *EventsHandler) eventWorker() {
	defer h.poolWg.Done()
	for {
		select {
		case work, ok := <-h.eventQueue:
			if !ok {
				return
			}
			start := time.Now()
			err := h.SendEventWithContext(work.ctx, work.eventNamespace, work.eventType, work.payload)
			if err != nil {
				slog.Error("Error sending async event", "namespace", work.eventNamespace, "type", work.eventType, "error", err)
			}
			if h.config.Hooks != nil && h.config.Hooks.OnEventProcessed != nil {
				h.config.Hooks.OnEventProcessed(work.eventNamespace, work.eventType, time.Since(start), err)
			}
		case <-h.poolStopCh:
			return
		}
	}
}

// SendEventAsync queues an event for asynchronous sending via the worker pool.
// Returns true if the event was queued, false if dropped (queue full or shutting down).
func (h *EventsHandler) SendEventAsync(eventNamespace string, eventType string, payload interface{}) bool {
	return h.SendEventAsyncWithContext(context.Background(), eventNamespace, eventType, payload)
}

// SendEventAsyncWithContext queues an event for asynchronous sending with context support.
// Returns true if the event was queued, false if dropped.
func (h *EventsHandler) SendEventAsyncWithContext(ctx context.Context, eventNamespace string, eventType string, payload interface{}) bool {
	h.shutdownMu.RLock()
	if h.shuttingDown {
		h.shutdownMu.RUnlock()
		if h.config.Hooks != nil && h.config.Hooks.OnEventQueued != nil {
			h.config.Hooks.OnEventQueued(eventNamespace, eventType, true)
		}
		slog.Warn("Event rejected, handler shutting down", "namespace", eventNamespace, "type", eventType)
		return false
	}
	h.shutdownMu.RUnlock()

	h.initWorkerPool()

	select {
	case h.eventQueue <- eventWork{ctx, eventNamespace, eventType, payload}:
		if h.config.Hooks != nil && h.config.Hooks.OnEventQueued != nil {
			h.config.Hooks.OnEventQueued(eventNamespace, eventType, false)
		}
		return true
	default:
		if h.config.Hooks != nil && h.config.Hooks.OnEventQueued != nil {
			h.config.Hooks.OnEventQueued(eventNamespace, eventType, true)
		}
		slog.Warn("Event queue full, dropping event", "namespace", eventNamespace, "type", eventType)
		return false
	}
}

// Shutdown gracefully stops the event worker pool, draining queued events.
// It blocks until all queued events are processed or the context is cancelled.
// After Shutdown is called, SendEventAsync will reject new events.
func (h *EventsHandler) Shutdown(ctx context.Context) error {
	var shutdownErr error
	h.shutdownOnce.Do(func() {
		// Mark as shutting down to reject new events
		h.shutdownMu.Lock()
		h.shuttingDown = true
		h.shutdownMu.Unlock()

		// If pool was never initialized, nothing to do
		if h.eventQueue == nil {
			return
		}

		slog.Info("Shutting down event handler, draining queue", "queuedEvents", len(h.eventQueue))

		// Close the queue to signal workers to drain and exit
		close(h.eventQueue)

		// Wait for workers to finish or context to cancel
		done := make(chan struct{})
		go func() {
			h.poolWg.Wait()
			close(done)
		}()

		select {
		case <-done:
			slog.Info("Event handler shutdown complete, all events processed")
		case <-ctx.Done():
			// Context cancelled, signal workers to stop immediately
			close(h.poolStopCh)
			shutdownErr = ctx.Err()
			slog.Warn("Event handler shutdown interrupted", "error", shutdownErr, "remainingEvents", len(h.eventQueue))
		}
	})
	return shutdownErr
}

// --- Global/package-level functions for backward compatibility ---

var (
	defaultHandler     *EventsHandler
	defaultHandlerOnce sync.Once
	defaultHandlerMu   sync.RWMutex

	// Legacy support: allow overriding config before first use
	pendingConfig   *Config
	pendingConfigMu sync.Mutex
)

// Init initializes the global events handler with the given configuration.
// If cfg is nil, DefaultConfig() is used.
// This function is safe to call multiple times; subsequent calls are no-ops.
func Init(cfg *Config) {
	pendingConfigMu.Lock()
	if cfg != nil {
		pendingConfig = cfg
	}
	pendingConfigMu.Unlock()
	getDefaultHandler()
}

func getDefaultHandler() *EventsHandler {
	defaultHandlerOnce.Do(func() {
		pendingConfigMu.Lock()
		cfg := pendingConfig
		pendingConfigMu.Unlock()

		if cfg == nil {
			c := DefaultConfig()
			cfg = &c
		}
		defaultHandler = NewEventsHandler(*cfg, nil)
	})
	return defaultHandler
}

// GetNatsClient returns the NATS client from the global handler, creating it if necessary.
func GetNatsClient() (NatsClient, error) {
	return getDefaultHandler().GetClient()
}

// SetNatsClient sets a custom NATS client on the global handler (useful for testing).
func SetNatsClient(client NatsClient) {
	getDefaultHandler().SetClient(client)
}

// ResetNatsClient resets the NATS client on the global handler.
func ResetNatsClient() {
	h := getDefaultHandler()
	h.clientMu.Lock()
	defer h.clientMu.Unlock()
	h.client = nil
	h.clientErr = nil
}

// SendEvent sends an event synchronously using the global handler.
func SendEvent(eventNamespace string, eventType string, payload interface{}) error {
	return getDefaultHandler().SendEvent(eventNamespace, eventType, payload)
}

// SendEventWithContext sends an event synchronously with context support using the global handler.
func SendEventWithContext(ctx context.Context, eventNamespace string, eventType string, payload interface{}) error {
	return getDefaultHandler().SendEventWithContext(ctx, eventNamespace, eventType, payload)
}

// SendEventWithClient sends an event using the provided client via the global handler.
func SendEventWithClient(ctx context.Context, client NatsClient, eventNamespace string, eventType string, payload interface{}) error {
	return getDefaultHandler().SendEventWithClient(ctx, client, eventNamespace, eventType, payload)
}

// SendEventAsync queues an event for asynchronous sending using the global handler.
func SendEventAsync(eventNamespace string, eventType string, payload interface{}) bool {
	return getDefaultHandler().SendEventAsync(eventNamespace, eventType, payload)
}

// SendEventAsyncWithContext queues an event for asynchronous sending with context using the global handler.
func SendEventAsyncWithContext(ctx context.Context, eventNamespace string, eventType string, payload interface{}) bool {
	return getDefaultHandler().SendEventAsyncWithContext(ctx, eventNamespace, eventType, payload)
}

// StopEventWorkerPool gracefully stops the global event worker pool.
// Deprecated: Use Shutdown(ctx) for graceful shutdown with queue draining.
func StopEventWorkerPool() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_ = getDefaultHandler().Shutdown(ctx)
}

// Shutdown gracefully stops the global event handler, draining queued events.
func Shutdown(ctx context.Context) error {
	return getDefaultHandler().Shutdown(ctx)
}

// GetConfig returns the current configuration from the global handler.
func GetConfig() Config {
	return getDefaultHandler().GetConfig()
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

// ResetDefaultHandler resets the global handler for testing purposes.
// This should only be used in tests.
func ResetDefaultHandler() {
	defaultHandlerMu.Lock()
	defer defaultHandlerMu.Unlock()
	defaultHandlerOnce = sync.Once{}
	defaultHandler = nil
	pendingConfigMu.Lock()
	pendingConfig = nil
	pendingConfigMu.Unlock()
}
