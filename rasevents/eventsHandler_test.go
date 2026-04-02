package rasevents

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	nats_service "github.com/transactrx/nats-service/pkg/nats-service"
	nats_service_client "github.com/transactrx/nats-service/pkg/nats-service-client"
)

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

// resetTestState resets global state between tests.
func resetTestState() {
	ResetNatsClient()
	// Reset config by creating a new one
	configOnce = sync.Once{}
	poolOnce = sync.Once{}
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

	err := SendEventWithClient(context.Background(), mockClient, "", eventType, payload)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	expectedSubject := "trx.eventscollector.collect.PatientNotification.TestType"
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
		{
			name:            "Default namespace",
			namespace:       "",
			eventType:       "TestType",
			expectedSubject: "trx.eventscollector.collect.PatientNotification.TestType",
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

	if cfg.DefaultNamespace != "PatientNotification" {
		t.Errorf("Expected DefaultNamespace 'PatientNotification', got: '%s'", cfg.DefaultNamespace)
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
