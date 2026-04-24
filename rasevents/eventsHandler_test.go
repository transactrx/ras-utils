package rasevents

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	nats_service "github.com/transactrx/nats-service/pkg/nats-service"
	nats_service_client "github.com/transactrx/nats-service/pkg/nats-service-client"
)

func TestMain(m *testing.M) {
	os.Setenv("EVENTS_DEFAULT_NAMESPACE", "TestNamespace")
	os.Setenv("EVENTS_SUBJECT", "trx.eventscollector.collect")
	code := m.Run()
	os.Unsetenv("EVENTS_DEFAULT_NAMESPACE")
	os.Unsetenv("EVENTS_SUBJECT")
	os.Exit(code)
}

// MockNatsClient is a mock NATS client for testing.
type MockNatsClient struct {
	DoRequestFunc func(correlationId, subject string, headers nats_service_client.Header, data []byte, timeout time.Duration) (*nats_service_client.NatsResponseMessage, *nats_service.NatsServiceError, error)
	callCount     int64 // Use atomic operations
	mu            sync.Mutex
	LastService   string
	LastSubject   string
	LastHeaders   map[string]string
	LastData      []byte
	LastTimeout   time.Duration
}

func (m *MockNatsClient) DoRequest(correlationId, subject string, headers nats_service_client.Header, data []byte, timeout time.Duration) (*nats_service_client.NatsResponseMessage, *nats_service.NatsServiceError, error) {
	atomic.AddInt64(&m.callCount, 1)

	m.mu.Lock()
	m.LastService = correlationId
	m.LastSubject = subject
	m.LastHeaders = map[string]string{}
	m.LastData = data
	m.LastTimeout = timeout
	m.mu.Unlock()

	if m.DoRequestFunc != nil {
		return m.DoRequestFunc(correlationId, subject, headers, data, timeout)
	}
	return &nats_service_client.NatsResponseMessage{Data: []byte("success")}, nil, nil
}

func (m *MockNatsClient) CallCount() int {
	return int(atomic.LoadInt64(&m.callCount))
}

func (m *MockNatsClient) ResetCallCount() {
	atomic.StoreInt64(&m.callCount, 0)
}

// resetTestState resets global state between tests and sets env vars
// so DefaultConfig() picks up the expected defaults.
func resetTestState() {
	ResetDefaultHandler()
	os.Setenv("EVENTS_DEFAULT_NAMESPACE", "TestNamespace")
	os.Setenv("EVENTS_SUBJECT", "trx.eventscollector.collect")
}

// TestCreateEvent_ValidPayload tests creating an event with valid payload
func TestCreateEvent_ValidPayload(t *testing.T) {
	payload := map[string]interface{}{
		"message": "test message",
		"id":      12345,
		"active":  true,
	}

	eventBytes, err := CreateEvent(payload)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(eventBytes) == 0 {
		t.Fatal("Expected non-empty event bytes")
	}

	var event Event
	err = json.Unmarshal(eventBytes, &event)
	if err != nil {
		t.Fatalf("Expected valid JSON, got error: %v", err)
	}

	if event.EventID == "" {
		t.Error("Expected non-empty EventID")
	}

	_, err = uuid.Parse(event.EventID)
	if err != nil {
		t.Errorf("Expected valid UUID for EventID, got: %s", event.EventID)
	}

	if event.EventTime == "" {
		t.Error("Expected non-empty EventTime")
	}

	_, err = time.Parse(time.RFC3339, event.EventTime)
	if err != nil {
		t.Errorf("Expected valid RFC3339 timestamp, got: %s", event.EventTime)
	}

	if event.EventPayload == nil {
		t.Error("Expected non-nil EventPayload")
	}

	payloadMap, ok := event.EventPayload.(map[string]interface{})
	if !ok {
		t.Error("Expected EventPayload to be a map")
	}

	if payloadMap["message"] != "test message" {
		t.Errorf("Expected message 'test message', got: %v", payloadMap["message"])
	}

	if payloadMap["id"] != float64(12345) {
		t.Errorf("Expected id 12345, got: %v", payloadMap["id"])
	}

	if payloadMap["active"] != true {
		t.Errorf("Expected active true, got: %v", payloadMap["active"])
	}
}

// TestCreateEvent_NilPayload tests creating an event with nil payload
func TestCreateEvent_NilPayload(t *testing.T) {
	eventBytes, err := CreateEvent(nil)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	var event Event
	err = json.Unmarshal(eventBytes, &event)
	if err != nil {
		t.Fatalf("Expected valid JSON, got error: %v", err)
	}

	if event.EventID == "" {
		t.Error("Expected non-empty EventID")
	}

	if event.EventTime == "" {
		t.Error("Expected non-empty EventTime")
	}

	if event.EventPayload != nil {
		t.Error("Expected nil EventPayload")
	}
}

// TestCreateEvent_StringPayload tests creating an event with string payload
func TestCreateEvent_StringPayload(t *testing.T) {
	payload := "test string payload"

	eventBytes, err := CreateEvent(payload)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	var event Event
	err = json.Unmarshal(eventBytes, &event)
	if err != nil {
		t.Fatalf("Expected valid JSON, got error: %v", err)
	}

	if event.EventPayload != payload {
		t.Errorf("Expected payload '%s', got: %v", payload, event.EventPayload)
	}
}

// TestCreateEvent_StructPayload tests creating an event with struct payload
func TestCreateEvent_StructPayload(t *testing.T) {
	type TestStruct struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}
	payload := TestStruct{
		Name:  "test",
		Value: 42,
	}

	eventBytes, err := CreateEvent(payload)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	var event Event
	err = json.Unmarshal(eventBytes, &event)
	if err != nil {
		t.Fatalf("Expected valid JSON, got error: %v", err)
	}

	payloadMap, ok := event.EventPayload.(map[string]interface{})
	if !ok {
		t.Error("Expected EventPayload to be a map")
	}

	if payloadMap["name"] != "test" {
		t.Errorf("Expected name 'test', got: %v", payloadMap["name"])
	}

	if payloadMap["value"] != float64(42) {
		t.Errorf("Expected value 42, got: %v", payloadMap["value"])
	}
}

// TestCreateEvent_SlicePayload tests creating an event with slice payload
func TestCreateEvent_SlicePayload(t *testing.T) {
	payload := []string{"item1", "item2", "item3"}

	eventBytes, err := CreateEvent(payload)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	var event Event
	err = json.Unmarshal(eventBytes, &event)
	if err != nil {
		t.Fatalf("Expected valid JSON, got error: %v", err)
	}

	payloadSlice, ok := event.EventPayload.([]interface{})
	if !ok {
		t.Error("Expected EventPayload to be a slice")
	}

	if len(payloadSlice) != 3 {
		t.Errorf("Expected 3 items, got: %d", len(payloadSlice))
	}

	for i, item := range payloadSlice {
		expected := fmt.Sprintf("item%d", i+1)
		if item != expected {
			t.Errorf("Expected item '%s', got: %v", expected, item)
		}
	}
}

// TestCreateEvent_UnmarshalablePayload tests creating an event with unmarshalable payload
func TestCreateEvent_UnmarshalablePayload(t *testing.T) {
	payload := make(chan int)

	eventBytes, err := CreateEvent(payload)

	if err == nil {
		t.Error("Expected error for unmarshalable payload")
	}

	if eventBytes != nil {
		t.Error("Expected nil bytes for error case")
	}
}

// TestCreateEvent_EmptyPayload tests creating an event with empty payload
func TestCreateEvent_EmptyPayload(t *testing.T) {
	payload := map[string]interface{}{}

	eventBytes, err := CreateEvent(payload)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	var event Event
	err = json.Unmarshal(eventBytes, &event)
	if err != nil {
		t.Fatalf("Expected valid JSON, got error: %v", err)
	}

	payloadMap, ok := event.EventPayload.(map[string]interface{})
	if !ok {
		t.Error("Expected EventPayload to be a map")
	}

	if len(payloadMap) != 0 {
		t.Errorf("Expected empty map, got: %v", payloadMap)
	}
}

// TestCreateEvent_UniqueEventIDs tests that each event gets a unique ID
func TestCreateEvent_UniqueEventIDs(t *testing.T) {
	payload := "test"
	eventIDs := make(map[string]bool)
	numEvents := 100

	for i := 0; i < numEvents; i++ {
		eventBytes, err := CreateEvent(payload)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		var event Event
		err = json.Unmarshal(eventBytes, &event)
		if err != nil {
			t.Fatalf("Expected valid JSON, got error: %v", err)
		}

		if eventIDs[event.EventID] {
			t.Errorf("Duplicate event ID found: %s", event.EventID)
		}
		eventIDs[event.EventID] = true
	}

	if len(eventIDs) != numEvents {
		t.Errorf("Expected %d unique IDs, got: %d", numEvents, len(eventIDs))
	}
}

// TestCreateEvent_TimestampProgression tests that timestamps progress over time
func TestCreateEvent_TimestampProgression(t *testing.T) {
	payload := "test"

	eventBytes1, err := CreateEvent(payload)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	time.Sleep(1 * time.Millisecond)

	eventBytes2, err := CreateEvent(payload)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	var event1, event2 Event
	if err := json.Unmarshal(eventBytes1, &event1); err != nil {
		t.Fatalf("Failed to unmarshal event1: %v", err)
	}
	if err := json.Unmarshal(eventBytes2, &event2); err != nil {
		t.Fatalf("Failed to unmarshal event2: %v", err)
	}

	time1, err := time.Parse(time.RFC3339, event1.EventTime)
	if err != nil {
		t.Fatalf("Expected valid timestamp, got error: %v", err)
	}

	time2, err := time.Parse(time.RFC3339, event2.EventTime)
	if err != nil {
		t.Fatalf("Expected valid timestamp, got error: %v", err)
	}

	if !time2.After(time1) && !time2.Equal(time1) {
		t.Errorf("Expected second event timestamp to be after first: %v vs %v", time1, time2)
	}
}

// TestSendEventWithClient_Success tests successful event sending
func TestSendEventWithClient_Success(t *testing.T) {
	resetTestState()

	mockClient := &MockNatsClient{
		DoRequestFunc: func(correlationId, subject string, headers nats_service_client.Header, data []byte, timeout time.Duration) (*nats_service_client.NatsResponseMessage, *nats_service.NatsServiceError, error) {
			return &nats_service_client.NatsResponseMessage{Data: []byte("success")}, nil, nil
		},
	}

	payload := map[string]string{"test": "data"}
	eventNamespace := "TestNamespace"
	eventType := "TestType"

	err := SendEventWithClient(context.Background(), mockClient, eventNamespace, eventType, payload)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if mockClient.CallCount() != 1 {
		t.Errorf("Expected 1 call to DoRequest, got: %d", mockClient.CallCount())
	}

	expectedSubject := "trx.eventscollector.collect.TestNamespace.TestType"
	if mockClient.LastSubject != expectedSubject {
		t.Errorf("Expected subject '%s', got: '%s'", expectedSubject, mockClient.LastSubject)
	}

	if mockClient.LastService != "" {
		t.Errorf("Expected empty service, got: '%s'", mockClient.LastService)
	}

	cfg := GetConfig()
	if mockClient.LastTimeout != cfg.Timeout {
		t.Errorf("Expected timeout of %v, got: %v", cfg.Timeout, mockClient.LastTimeout)
	}

	if len(mockClient.LastData) == 0 {
		t.Error("Expected non-empty data")
	}

	var event Event
	err = json.Unmarshal(mockClient.LastData, &event)
	if err != nil {
		t.Errorf("Expected valid JSON event, got error: %v", err)
	}
}

// TestSendEventWithClient_DefaultNamespace tests sending event with default namespace
func TestSendEventWithClient_DefaultNamespace(t *testing.T) {
	resetTestState()

	mockClient := &MockNatsClient{
		DoRequestFunc: func(correlationId, subject string, headers nats_service_client.Header, data []byte, timeout time.Duration) (*nats_service_client.NatsResponseMessage, *nats_service.NatsServiceError, error) {
			return &nats_service_client.NatsResponseMessage{Data: []byte("success")}, nil, nil
		},
	}

	payload := map[string]string{"test": "data"}
	eventType := "TestType"

	err := SendEventWithClient(context.Background(), mockClient, "TestNamespace", eventType, payload)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	expectedSubject := "trx.eventscollector.collect.TestNamespace.TestType"
	if mockClient.LastSubject != expectedSubject {
		t.Errorf("Expected subject '%s', got: '%s'", expectedSubject, mockClient.LastSubject)
	}
}

// TestSendEventWithClient_NatsError tests handling NATS service errors
func TestSendEventWithClient_NatsError(t *testing.T) {
	resetTestState()

	natsServiceError := &nats_service.NatsServiceError{
		Status:       500,
		ErrorMessage: "Connection failed",
	}

	mockClient := &MockNatsClient{
		DoRequestFunc: func(correlationId, subject string, headers nats_service_client.Header, data []byte, timeout time.Duration) (*nats_service_client.NatsResponseMessage, *nats_service.NatsServiceError, error) {
			return nil, natsServiceError, nil
		},
	}

	payload := map[string]string{"test": "data"}

	err := SendEventWithClient(context.Background(), mockClient, "TestNamespace", "TestType", payload)

	if err == nil {
		t.Error("Expected error for NATS service error")
	}

	if !strings.Contains(err.Error(), "NATS service error") {
		t.Errorf("Expected error message to contain 'NATS service error', got: %v", err)
	}
}

// TestSendEventWithClient_RequestError tests handling request errors
func TestSendEventWithClient_RequestError(t *testing.T) {
	resetTestState()

	requestError := errors.New("request failed")

	mockClient := &MockNatsClient{
		DoRequestFunc: func(correlationId, subject string, headers nats_service_client.Header, data []byte, timeout time.Duration) (*nats_service_client.NatsResponseMessage, *nats_service.NatsServiceError, error) {
			return nil, nil, requestError
		},
	}

	payload := map[string]string{"test": "data"}

	err := SendEventWithClient(context.Background(), mockClient, "TestNamespace", "TestType", payload)

	if err == nil {
		t.Error("Expected error for request error")
	}

	if !strings.Contains(err.Error(), "request failed") {
		t.Errorf("Expected error message to contain 'request failed', got: %v", err)
	}
}

// TestSendEventWithClient_InvalidPayload tests handling invalid payload
func TestSendEventWithClient_InvalidPayload(t *testing.T) {
	resetTestState()

	mockClient := &MockNatsClient{}
	payload := make(chan int)

	err := SendEventWithClient(context.Background(), mockClient, "TestNamespace", "TestType", payload)

	if err == nil {
		t.Error("Expected error for invalid payload")
	}

	if !strings.Contains(err.Error(), "error creating event") {
		t.Errorf("Expected error message to contain 'error creating event', got: %v", err)
	}

	if mockClient.CallCount() != 0 {
		t.Errorf("Expected 0 calls to DoRequest for invalid payload, got: %d", mockClient.CallCount())
	}
}

// TestSendEventWithClient_ContextCancellation tests context cancellation
func TestSendEventWithClient_ContextCancellation(t *testing.T) {
	resetTestState()

	mockClient := &MockNatsClient{}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	payload := map[string]string{"test": "data"}

	err := SendEventWithClient(ctx, mockClient, "TestNamespace", "TestType", payload)

	if err == nil {
		t.Error("Expected error for cancelled context")
	}

	if !errors.Is(err, context.Canceled) {
		t.Errorf("Expected context.Canceled error, got: %v", err)
	}

	if mockClient.CallCount() != 0 {
		t.Errorf("Expected 0 calls to DoRequest for cancelled context, got: %d", mockClient.CallCount())
	}
}

// TestSendEventWithClient_SubjectConstruction tests various subject constructions
func TestSendEventWithClient_SubjectConstruction(t *testing.T) {
	testCases := []struct {
		name            string
		namespace       string
		eventType       string
		expectedSubject string
	}{
		{
			name:            "Normal case",
			namespace:       "PatientNotification",
			eventType:       "Email",
			expectedSubject: "trx.eventscollector.collect.PatientNotification.Email",
		},
		{
			name:            "SMS notification",
			namespace:       "PatientNotification",
			eventType:       "SMS",
			expectedSubject: "trx.eventscollector.collect.PatientNotification.SMS",
		},
		{
			name:            "Audit event",
			namespace:       "ClinicalPlus",
			eventType:       "Audit",
			expectedSubject: "trx.eventscollector.collect.ClinicalPlus.Audit",
		},
		{
			name:            "Custom namespace",
			namespace:       "CustomService",
			eventType:       "CustomEvent",
			expectedSubject: "trx.eventscollector.collect.CustomService.CustomEvent",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resetTestState()

			mockClient := &MockNatsClient{
				DoRequestFunc: func(correlationId, subject string, headers nats_service_client.Header, data []byte, timeout time.Duration) (*nats_service_client.NatsResponseMessage, *nats_service.NatsServiceError, error) {
					return &nats_service_client.NatsResponseMessage{Data: []byte("success")}, nil, nil
				},
			}

			payload := map[string]string{"test": "data"}

			err := SendEventWithClient(context.Background(), mockClient, tc.namespace, tc.eventType, payload)

			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}

			if mockClient.LastSubject != tc.expectedSubject {
				t.Errorf("Expected subject '%s', got: '%s'", tc.expectedSubject, mockClient.LastSubject)
			}
		})
	}
}

// TestSendEventWithClient_ConcurrentCalls tests concurrent event sending
func TestSendEventWithClient_ConcurrentCalls(t *testing.T) {
	resetTestState()

	mockClient := &MockNatsClient{
		DoRequestFunc: func(correlationId, subject string, headers nats_service_client.Header, data []byte, timeout time.Duration) (*nats_service_client.NatsResponseMessage, *nats_service.NatsServiceError, error) {
			return &nats_service_client.NatsResponseMessage{Data: []byte("success")}, nil, nil
		},
	}

	payload := map[string]string{"test": "data"}
	numCalls := 50

	var wg sync.WaitGroup
	errChan := make(chan error, numCalls)

	for i := 0; i < numCalls; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			err := SendEventWithClient(context.Background(), mockClient, "TestNamespace", fmt.Sprintf("TestType%d", index), payload)
			if err != nil {
				errChan <- err
			}
		}(i)
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		t.Errorf("Unexpected error: %v", err)
	}

	if mockClient.CallCount() != numCalls {
		t.Errorf("Expected %d calls to DoRequest, got: %d", numCalls, mockClient.CallCount())
	}
}

// TestConfig_Defaults tests default configuration values
func TestConfig_Defaults(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.DefaultNamespace != "TestNamespace" {
		t.Errorf("Expected DefaultNamespace 'TestNamespace', got: '%s'", cfg.DefaultNamespace)
	}

	if cfg.Subject != "trx.eventscollector.collect" {
		t.Errorf("Expected Subject 'trx.eventscollector.collect', got: '%s'", cfg.Subject)
	}

	if cfg.Timeout != 60*time.Second {
		t.Errorf("Expected Timeout 60s, got: %v", cfg.Timeout)
	}

	if cfg.WorkerPoolSize != 50 {
		t.Errorf("Expected WorkerPoolSize 50, got: %d", cfg.WorkerPoolSize)
	}

	if cfg.EventQueueSize != 1000 {
		t.Errorf("Expected EventQueueSize 1000, got: %d", cfg.EventQueueSize)
	}
}

// TestInit_CustomConfig tests initialization with custom configuration
func TestInit_CustomConfig(t *testing.T) {
	resetTestState()

	customCfg := &Config{
		DefaultNamespace: "CustomNamespace",
		Subject:          "custom.subject",
		Timeout:          30 * time.Second,
		WorkerPoolSize:   10,
		EventQueueSize:   100,
	}

	Init(customCfg)

	cfg := GetConfig()

	if cfg.DefaultNamespace != "CustomNamespace" {
		t.Errorf("Expected DefaultNamespace 'CustomNamespace', got: '%s'", cfg.DefaultNamespace)
	}

	if cfg.Subject != "custom.subject" {
		t.Errorf("Expected Subject 'custom.subject', got: '%s'", cfg.Subject)
	}

	if cfg.Timeout != 30*time.Second {
		t.Errorf("Expected Timeout 30s, got: %v", cfg.Timeout)
	}
}

// TestSetNatsClient tests setting a custom NATS client
func TestSetNatsClient(t *testing.T) {
	resetTestState()

	mockClient := &MockNatsClient{}
	SetNatsClient(mockClient)

	client, err := GetNatsClient()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if client != mockClient {
		t.Error("Expected to get the mock client back")
	}
}

// TestSendEvent_WithSetClient tests SendEvent using SetNatsClient
func TestSendEvent_WithSetClient(t *testing.T) {
	resetTestState()

	mockClient := &MockNatsClient{
		DoRequestFunc: func(correlationId, subject string, headers nats_service_client.Header, data []byte, timeout time.Duration) (*nats_service_client.NatsResponseMessage, *nats_service.NatsServiceError, error) {
			return &nats_service_client.NatsResponseMessage{Data: []byte("success")}, nil, nil
		},
	}
	SetNatsClient(mockClient)

	err := SendEvent("TestNamespace", "TestType", map[string]string{"test": "data"})

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if mockClient.CallCount() != 1 {
		t.Errorf("Expected 1 call, got: %d", mockClient.CallCount())
	}
}

// TestSendEventAsync_Success tests async event sending
func TestSendEventAsync_Success(t *testing.T) {
	resetTestState()

	var callCount int64
	mockClient := &MockNatsClient{
		DoRequestFunc: func(correlationId, subject string, headers nats_service_client.Header, data []byte, timeout time.Duration) (*nats_service_client.NatsResponseMessage, *nats_service.NatsServiceError, error) {
			atomic.AddInt64(&callCount, 1)
			return &nats_service_client.NatsResponseMessage{Data: []byte("success")}, nil, nil
		},
	}
	SetNatsClient(mockClient)

	queued := SendEventAsync("TestNamespace", "TestType", map[string]string{"test": "data"})

	if !queued {
		t.Error("Expected event to be queued")
	}

	// Give the worker time to process
	time.Sleep(100 * time.Millisecond)

	if atomic.LoadInt64(&callCount) != 1 {
		t.Errorf("Expected 1 async call, got: %d", atomic.LoadInt64(&callCount))
	}
}

// TestSendEventAsync_ReturnsStatus tests that SendEventAsync returns queue status
func TestSendEventAsync_ReturnsStatus(t *testing.T) {
	resetTestState()

	mockClient := &MockNatsClient{}
	SetNatsClient(mockClient)

	// First call should succeed
	queued := SendEventAsync("TestNamespace", "TestType", map[string]string{"test": "data"})
	if !queued {
		t.Error("Expected first event to be queued")
	}
}

// TestEventStruct tests the Event struct
func TestEventStruct(t *testing.T) {
	event := Event{
		EventID:      "test-id",
		EventTime:    "2023-01-01T00:00:00Z",
		EventPayload: "test-payload",
	}

	if event.EventID != "test-id" {
		t.Errorf("Expected EventID 'test-id', got: '%s'", event.EventID)
	}

	if event.EventTime != "2023-01-01T00:00:00Z" {
		t.Errorf("Expected EventTime '2023-01-01T00:00:00Z', got: '%s'", event.EventTime)
	}

	if event.EventPayload != "test-payload" {
		t.Errorf("Expected EventPayload 'test-payload', got: '%v'", event.EventPayload)
	}
}

// TestEventStruct_JSONTags tests the Event struct JSON tags
func TestEventStruct_JSONTags(t *testing.T) {
	event := Event{
		EventID:      "test-id",
		EventTime:    "2023-01-01T00:00:00Z",
		EventPayload: "test-payload",
	}

	jsonBytes, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Expected no error marshaling event, got: %v", err)
	}

	var jsonMap map[string]interface{}
	err = json.Unmarshal(jsonBytes, &jsonMap)
	if err != nil {
		t.Fatalf("Expected no error unmarshaling JSON, got: %v", err)
	}

	if jsonMap["eventId"] != "test-id" {
		t.Errorf("Expected eventId field in JSON, got: %v", jsonMap)
	}

	if jsonMap["eventTime"] != "2023-01-01T00:00:00Z" {
		t.Errorf("Expected eventTime field in JSON, got: %v", jsonMap)
	}

	if jsonMap["eventPayload"] != "test-payload" {
		t.Errorf("Expected eventPayload field in JSON, got: %v", jsonMap)
	}
}

// --- New tests for EventsHandler struct ---

// TestNewEventsHandler tests creating a new EventsHandler
func TestNewEventsHandler(t *testing.T) {
	cfg := Config{
		DefaultNamespace: "TestNamespace",
		Subject:          "test.subject",
		Timeout:          30 * time.Second,
		WorkerPoolSize:   5,
		EventQueueSize:   10,
	}

	mockClient := &MockNatsClient{}
	handler := NewEventsHandler(cfg, mockClient)

	if handler == nil {
		t.Fatal("Expected non-nil handler")
	}

	gotCfg := handler.GetConfig()
	if gotCfg.DefaultNamespace != "TestNamespace" {
		t.Errorf("Expected DefaultNamespace 'TestNamespace', got: '%s'", gotCfg.DefaultNamespace)
	}

	client, err := handler.GetClient()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if client != mockClient {
		t.Error("Expected to get the mock client back")
	}
}

// TestEventsHandler_SendEvent tests sending events via handler instance
func TestEventsHandler_SendEvent(t *testing.T) {
	mockClient := &MockNatsClient{
		DoRequestFunc: func(correlationId, subject string, headers nats_service_client.Header, data []byte, timeout time.Duration) (*nats_service_client.NatsResponseMessage, *nats_service.NatsServiceError, error) {
			return &nats_service_client.NatsResponseMessage{Data: []byte("success")}, nil, nil
		},
	}

	cfg := Config{
		DefaultNamespace: "HandlerNamespace",
		Subject:          "handler.subject",
		Timeout:          10 * time.Second,
		WorkerPoolSize:   2,
		EventQueueSize:   10,
	}

	handler := NewEventsHandler(cfg, mockClient)
	err := handler.SendEvent("TestNS", "TestType", map[string]string{"key": "value"})

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	expectedSubject := "handler.subject.TestNS.TestType"
	if mockClient.LastSubject != expectedSubject {
		t.Errorf("Expected subject '%s', got: '%s'", expectedSubject, mockClient.LastSubject)
	}
}

// TestEventsHandler_SendEventAsync tests async sending via handler instance
func TestEventsHandler_SendEventAsync(t *testing.T) {
	var callCount int64
	mockClient := &MockNatsClient{
		DoRequestFunc: func(correlationId, subject string, headers nats_service_client.Header, data []byte, timeout time.Duration) (*nats_service_client.NatsResponseMessage, *nats_service.NatsServiceError, error) {
			atomic.AddInt64(&callCount, 1)
			return &nats_service_client.NatsResponseMessage{Data: []byte("success")}, nil, nil
		},
	}

	cfg := Config{
		DefaultNamespace: "AsyncNS",
		Subject:          "async.subject",
		Timeout:          10 * time.Second,
		WorkerPoolSize:   2,
		EventQueueSize:   10,
	}

	handler := NewEventsHandler(cfg, mockClient)
	queued := handler.SendEventAsync("TestNS", "TestType", map[string]string{"key": "value"})

	if !queued {
		t.Error("Expected event to be queued")
	}

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	if atomic.LoadInt64(&callCount) != 1 {
		t.Errorf("Expected 1 call, got: %d", atomic.LoadInt64(&callCount))
	}

	// Cleanup
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_ = handler.Shutdown(ctx)
}

// TestEventsHandler_Shutdown_DrainsQueue tests that Shutdown drains the queue
func TestEventsHandler_Shutdown_DrainsQueue(t *testing.T) {
	var callCount int64
	var processedMu sync.Mutex
	var processed []string

	mockClient := &MockNatsClient{
		DoRequestFunc: func(correlationId, subject string, headers nats_service_client.Header, data []byte, timeout time.Duration) (*nats_service_client.NatsResponseMessage, *nats_service.NatsServiceError, error) {
			atomic.AddInt64(&callCount, 1)
			// Simulate some work
			time.Sleep(10 * time.Millisecond)
			processedMu.Lock()
			processed = append(processed, subject)
			processedMu.Unlock()
			return &nats_service_client.NatsResponseMessage{Data: []byte("success")}, nil, nil
		},
	}

	cfg := Config{
		DefaultNamespace: "DrainNS",
		Subject:          "drain.subject",
		Timeout:          10 * time.Second,
		WorkerPoolSize:   2,
		EventQueueSize:   100,
	}

	handler := NewEventsHandler(cfg, mockClient)

	// Queue several events
	numEvents := 10
	for i := 0; i < numEvents; i++ {
		queued := handler.SendEventAsync("TestNS", fmt.Sprintf("Type%d", i), map[string]string{"key": "value"})
		if !queued {
			t.Errorf("Expected event %d to be queued", i)
		}
	}

	// Shutdown with enough time to drain
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := handler.Shutdown(ctx)
	if err != nil {
		t.Errorf("Expected no error from Shutdown, got: %v", err)
	}

	// All events should have been processed
	if atomic.LoadInt64(&callCount) != int64(numEvents) {
		t.Errorf("Expected %d events processed, got: %d", numEvents, atomic.LoadInt64(&callCount))
	}
}

// TestEventsHandler_Shutdown_RejectsNewEvents tests that Shutdown rejects new events
func TestEventsHandler_Shutdown_RejectsNewEvents(t *testing.T) {
	mockClient := &MockNatsClient{
		DoRequestFunc: func(correlationId, subject string, headers nats_service_client.Header, data []byte, timeout time.Duration) (*nats_service_client.NatsResponseMessage, *nats_service.NatsServiceError, error) {
			time.Sleep(50 * time.Millisecond) // Slow processing
			return &nats_service_client.NatsResponseMessage{Data: []byte("success")}, nil, nil
		},
	}

	cfg := Config{
		DefaultNamespace: "RejectNS",
		Subject:          "reject.subject",
		Timeout:          10 * time.Second,
		WorkerPoolSize:   1,
		EventQueueSize:   10,
	}

	handler := NewEventsHandler(cfg, mockClient)

	// Queue one event to initialize the pool
	handler.SendEventAsync("TestNS", "TestType", map[string]string{"key": "value"})

	// Start shutdown in background
	shutdownDone := make(chan struct{})
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = handler.Shutdown(ctx)
		close(shutdownDone)
	}()

	// Give shutdown time to set shuttingDown flag
	time.Sleep(10 * time.Millisecond)

	// Try to queue another event - should be rejected
	queued := handler.SendEventAsync("TestNS", "TestType2", map[string]string{"key": "value"})
	if queued {
		t.Error("Expected event to be rejected after Shutdown called")
	}

	<-shutdownDone
}

// TestEventsHandler_Shutdown_ContextTimeout tests Shutdown with context timeout
func TestEventsHandler_Shutdown_ContextTimeout(t *testing.T) {
	mockClient := &MockNatsClient{
		DoRequestFunc: func(correlationId, subject string, headers nats_service_client.Header, data []byte, timeout time.Duration) (*nats_service_client.NatsResponseMessage, *nats_service.NatsServiceError, error) {
			time.Sleep(500 * time.Millisecond) // Very slow processing
			return &nats_service_client.NatsResponseMessage{Data: []byte("success")}, nil, nil
		},
	}

	cfg := Config{
		DefaultNamespace: "TimeoutNS",
		Subject:          "timeout.subject",
		Timeout:          10 * time.Second,
		WorkerPoolSize:   1,
		EventQueueSize:   100,
	}

	handler := NewEventsHandler(cfg, mockClient)

	// Queue many events
	for i := 0; i < 20; i++ {
		handler.SendEventAsync("TestNS", fmt.Sprintf("Type%d", i), map[string]string{"key": "value"})
	}

	// Shutdown with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := handler.Shutdown(ctx)
	if err == nil {
		t.Error("Expected context deadline exceeded error")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Expected context.DeadlineExceeded, got: %v", err)
	}
}

// --- Tests for Hooks ---

// TestHooks_OnEventSent tests the OnEventSent hook
func TestHooks_OnEventSent(t *testing.T) {
	var hookCalled bool
	var hookNamespace, hookEventType string
	var hookDuration time.Duration
	var hookErr error

	mockClient := &MockNatsClient{
		DoRequestFunc: func(correlationId, subject string, headers nats_service_client.Header, data []byte, timeout time.Duration) (*nats_service_client.NatsResponseMessage, *nats_service.NatsServiceError, error) {
			time.Sleep(10 * time.Millisecond)
			return &nats_service_client.NatsResponseMessage{Data: []byte("success")}, nil, nil
		},
	}

	cfg := Config{
		DefaultNamespace: "HookNS",
		Subject:          "hook.subject",
		Timeout:          10 * time.Second,
		WorkerPoolSize:   2,
		EventQueueSize:   10,
		Hooks: &Hooks{
			OnEventSent: func(namespace, eventType string, duration time.Duration, err error) {
				hookCalled = true
				hookNamespace = namespace
				hookEventType = eventType
				hookDuration = duration
				hookErr = err
			},
		},
	}

	handler := NewEventsHandler(cfg, mockClient)
	err := handler.SendEvent("TestNS", "TestType", map[string]string{"key": "value"})

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if !hookCalled {
		t.Error("Expected OnEventSent hook to be called")
	}
	if hookNamespace != "TestNS" {
		t.Errorf("Expected namespace 'TestNS', got: '%s'", hookNamespace)
	}
	if hookEventType != "TestType" {
		t.Errorf("Expected eventType 'TestType', got: '%s'", hookEventType)
	}
	if hookDuration < 10*time.Millisecond {
		t.Errorf("Expected duration >= 10ms, got: %v", hookDuration)
	}
	if hookErr != nil {
		t.Errorf("Expected nil error in hook, got: %v", hookErr)
	}
}

// TestHooks_OnEventSent_Error tests OnEventSent hook on error
func TestHooks_OnEventSent_Error(t *testing.T) {
	var hookErr error

	mockClient := &MockNatsClient{
		DoRequestFunc: func(correlationId, subject string, headers nats_service_client.Header, data []byte, timeout time.Duration) (*nats_service_client.NatsResponseMessage, *nats_service.NatsServiceError, error) {
			return nil, nil, errors.New("send failed")
		},
	}

	cfg := Config{
		DefaultNamespace: "HookNS",
		Subject:          "hook.subject",
		Timeout:          10 * time.Second,
		WorkerPoolSize:   2,
		EventQueueSize:   10,
		Hooks: &Hooks{
			OnEventSent: func(namespace, eventType string, duration time.Duration, err error) {
				hookErr = err
			},
		},
	}

	handler := NewEventsHandler(cfg, mockClient)
	_ = handler.SendEvent("TestNS", "TestType", map[string]string{"key": "value"})

	if hookErr == nil {
		t.Error("Expected error to be passed to hook")
	}
	if !strings.Contains(hookErr.Error(), "send failed") {
		t.Errorf("Expected error to contain 'send failed', got: %v", hookErr)
	}
}

// TestHooks_OnEventQueued tests the OnEventQueued hook
func TestHooks_OnEventQueued(t *testing.T) {
	var hookCalled bool
	var hookNamespace, hookEventType string
	var hookDropped bool

	mockClient := &MockNatsClient{}

	cfg := Config{
		DefaultNamespace: "HookNS",
		Subject:          "hook.subject",
		Timeout:          10 * time.Second,
		WorkerPoolSize:   2,
		EventQueueSize:   10,
		Hooks: &Hooks{
			OnEventQueued: func(namespace, eventType string, dropped bool) {
				hookCalled = true
				hookNamespace = namespace
				hookEventType = eventType
				hookDropped = dropped
			},
		},
	}

	handler := NewEventsHandler(cfg, mockClient)
	queued := handler.SendEventAsync("TestNS", "TestType", map[string]string{"key": "value"})

	if !queued {
		t.Error("Expected event to be queued")
	}
	if !hookCalled {
		t.Error("Expected OnEventQueued hook to be called")
	}
	if hookNamespace != "TestNS" {
		t.Errorf("Expected namespace 'TestNS', got: '%s'", hookNamespace)
	}
	if hookEventType != "TestType" {
		t.Errorf("Expected eventType 'TestType', got: '%s'", hookEventType)
	}
	if hookDropped {
		t.Error("Expected dropped to be false")
	}

	// Cleanup
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_ = handler.Shutdown(ctx)
}

// TestHooks_OnEventQueued_Dropped tests OnEventQueued when queue is full
func TestHooks_OnEventQueued_Dropped(t *testing.T) {
	var droppedCount int64

	mockClient := &MockNatsClient{
		DoRequestFunc: func(correlationId, subject string, headers nats_service_client.Header, data []byte, timeout time.Duration) (*nats_service_client.NatsResponseMessage, *nats_service.NatsServiceError, error) {
			time.Sleep(100 * time.Millisecond) // Slow processing
			return &nats_service_client.NatsResponseMessage{Data: []byte("success")}, nil, nil
		},
	}

	cfg := Config{
		DefaultNamespace: "HookNS",
		Subject:          "hook.subject",
		Timeout:          10 * time.Second,
		WorkerPoolSize:   1,
		EventQueueSize:   2, // Very small queue
		Hooks: &Hooks{
			OnEventQueued: func(namespace, eventType string, dropped bool) {
				if dropped {
					atomic.AddInt64(&droppedCount, 1)
				}
			},
		},
	}

	handler := NewEventsHandler(cfg, mockClient)

	// Fill queue and overflow
	for i := 0; i < 10; i++ {
		handler.SendEventAsync("TestNS", fmt.Sprintf("Type%d", i), map[string]string{"key": "value"})
	}

	// Should have dropped some events
	if atomic.LoadInt64(&droppedCount) == 0 {
		t.Error("Expected some events to be dropped")
	}

	// Cleanup
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = handler.Shutdown(ctx)
}

// TestHooks_OnEventProcessed tests the OnEventProcessed hook
func TestHooks_OnEventProcessed(t *testing.T) {
	var hookCalled int64
	var lastNamespace, lastEventType string
	var lastDuration time.Duration
	var lastErr error
	var mu sync.Mutex

	mockClient := &MockNatsClient{
		DoRequestFunc: func(correlationId, subject string, headers nats_service_client.Header, data []byte, timeout time.Duration) (*nats_service_client.NatsResponseMessage, *nats_service.NatsServiceError, error) {
			time.Sleep(5 * time.Millisecond)
			return &nats_service_client.NatsResponseMessage{Data: []byte("success")}, nil, nil
		},
	}

	cfg := Config{
		DefaultNamespace: "HookNS",
		Subject:          "hook.subject",
		Timeout:          10 * time.Second,
		WorkerPoolSize:   2,
		EventQueueSize:   10,
		Hooks: &Hooks{
			OnEventProcessed: func(namespace, eventType string, duration time.Duration, err error) {
				atomic.AddInt64(&hookCalled, 1)
				mu.Lock()
				lastNamespace = namespace
				lastEventType = eventType
				lastDuration = duration
				lastErr = err
				mu.Unlock()
			},
		},
	}

	handler := NewEventsHandler(cfg, mockClient)
	handler.SendEventAsync("TestNS", "TestType", map[string]string{"key": "value"})

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	if atomic.LoadInt64(&hookCalled) != 1 {
		t.Errorf("Expected OnEventProcessed to be called once, got: %d", atomic.LoadInt64(&hookCalled))
	}

	mu.Lock()
	if lastNamespace != "TestNS" {
		t.Errorf("Expected namespace 'TestNS', got: '%s'", lastNamespace)
	}
	if lastEventType != "TestType" {
		t.Errorf("Expected eventType 'TestType', got: '%s'", lastEventType)
	}
	if lastDuration < 5*time.Millisecond {
		t.Errorf("Expected duration >= 5ms, got: %v", lastDuration)
	}
	if lastErr != nil {
		t.Errorf("Expected nil error, got: %v", lastErr)
	}
	mu.Unlock()

	// Cleanup
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_ = handler.Shutdown(ctx)
}

// TestMultipleHandlers tests using multiple independent handlers
func TestMultipleHandlers(t *testing.T) {
	mockClient1 := &MockNatsClient{}
	mockClient2 := &MockNatsClient{}

	cfg1 := Config{
		DefaultNamespace: "NS1",
		Subject:          "subject1",
		Timeout:          10 * time.Second,
		WorkerPoolSize:   2,
		EventQueueSize:   10,
	}

	cfg2 := Config{
		DefaultNamespace: "NS2",
		Subject:          "subject2",
		Timeout:          20 * time.Second,
		WorkerPoolSize:   4,
		EventQueueSize:   20,
	}

	handler1 := NewEventsHandler(cfg1, mockClient1)
	handler2 := NewEventsHandler(cfg2, mockClient2)

	// Send events through both handlers
	_ = handler1.SendEvent("", "Type1", map[string]string{"handler": "1"})
	_ = handler2.SendEvent("", "Type2", map[string]string{"handler": "2"})

	// Verify they used their own configs
	if mockClient1.LastSubject != "subject1.NS1.Type1" {
		t.Errorf("Expected subject 'subject1.NS1.Type1', got: '%s'", mockClient1.LastSubject)
	}
	if mockClient2.LastSubject != "subject2.NS2.Type2" {
		t.Errorf("Expected subject 'subject2.NS2.Type2', got: '%s'", mockClient2.LastSubject)
	}

	// Verify call counts are independent
	if mockClient1.CallCount() != 1 {
		t.Errorf("Expected handler1 to have 1 call, got: %d", mockClient1.CallCount())
	}
	if mockClient2.CallCount() != 1 {
		t.Errorf("Expected handler2 to have 1 call, got: %d", mockClient2.CallCount())
	}
}
