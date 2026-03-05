package anypointclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/pkg/errors"
)

// MqQueue represents an Anypoint MQ queue
type MqQueue struct {
	QueueID              string `json:"queueId,omitempty"`
	Type                 string `json:"type,omitempty"` // Always "queue" for queues
	Fifo                 bool   `json:"fifo,omitempty"`
	Encrypted            *bool  `json:"encrypted,omitempty"` // Defaults to true if not specified
	MaxDeliveries        int    `json:"maxDeliveries,omitempty"`
	DeadLetterQueueID    string `json:"deadLetterQueueId,omitempty"`
	IsFallback           bool   `json:"isFallback,omitempty"`
	DefaultTtl           int64  `json:"defaultTtl,omitempty"`
	DefaultLockTtl       int64  `json:"defaultLockTtl,omitempty"`
	DefaultDeliveryDelay int64  `json:"defaultDeliveryDelay,omitempty"`
}

// IsEncrypted returns the encrypted value, defaulting to true if not set
func (q *MqQueue) IsEncrypted() bool {
	if q.Encrypted == nil {
		return true
	}
	return *q.Encrypted
}

// Validate checks that queue settings are within valid ranges
// All time values are in milliseconds
func (q *MqQueue) Validate() error {
	// Max Deliveries (delivery attempts): 1 to 1000
	if q.MaxDeliveries != 0 && (q.MaxDeliveries < 1 || q.MaxDeliveries > 1000) {
		return fmt.Errorf("queue %s: maxDeliveries must be between 1 and 1000, got %d", q.QueueID, q.MaxDeliveries)
	}

	// Message TTL: 60000ms (1 min) to 1209600000ms (14 days)
	if q.DefaultTtl != 0 && (q.DefaultTtl < 60000 || q.DefaultTtl > 1209600000) {
		return fmt.Errorf("queue %s: defaultTtl must be between 60000 and 1209600000 ms (1 minute to 14 days), got %d", q.QueueID, q.DefaultTtl)
	}

	// Acknowledgement Timeout (Lock TTL): 0 to 43200000ms (12 hours)
	if q.DefaultLockTtl < 0 || q.DefaultLockTtl > 43200000 {
		return fmt.Errorf("queue %s: defaultLockTtl must be between 0 and 43200000 ms (0 to 12 hours), got %d", q.QueueID, q.DefaultLockTtl)
	}

	// Delivery Delay: 0 or 1000ms to 900000ms (1 second to 15 minutes)
	if q.DefaultDeliveryDelay != 0 && (q.DefaultDeliveryDelay < 1000 || q.DefaultDeliveryDelay > 900000) {
		return fmt.Errorf("queue %s: defaultDeliveryDelay must be 0 or between 1000 and 900000 ms (1 second to 15 minutes), got %d", q.QueueID, q.DefaultDeliveryDelay)
	}

	return nil
}

// mqQueueCreateRequest is the request body for creating a queue (queueId is in URL, not body)
type mqQueueCreateRequest struct {
	Type                 string `json:"type,omitempty"`
	Fifo                 bool   `json:"fifo,omitempty"`
	Encrypted            bool   `json:"encrypted"`
	MaxDeliveries        int    `json:"maxDeliveries,omitempty"`
	DeadLetterQueueID    string `json:"deadLetterQueueId,omitempty"`
	DefaultTtl           int64  `json:"defaultTtl,omitempty"`
	DefaultLockTtl       int64  `json:"defaultLockTtl,omitempty"`
	DefaultDeliveryDelay int64  `json:"defaultDeliveryDelay"`
}

// mqQueueUpdateRequest is the request body for updating a queue (uses PATCH)
type mqQueueUpdateRequest struct {
	Type                 string `json:"type"`
	Fifo                 bool   `json:"fifo"`
	Encrypted            bool   `json:"encrypted"`
	MaxDeliveries        int    `json:"maxDeliveries,omitempty"`
	DeadLetterQueueID    string `json:"deadLetterQueueId,omitempty"`
	IsFallback           bool   `json:"isFallback"`
	DefaultTtl           int64  `json:"defaultTtl,omitempty"`
	DefaultLockTtl       int64  `json:"defaultLockTtl,omitempty"`
	DefaultDeliveryDelay int64  `json:"defaultDeliveryDelay"`
}

// MqExchange represents an Anypoint MQ exchange
type MqExchange struct {
	ExchangeID string `json:"exchangeId"`
	Fifo       bool   `json:"fifo,omitempty"`
	Encrypted  *bool  `json:"encrypted,omitempty"` // Defaults to true if not specified
}

// IsEncrypted returns the encrypted value, defaulting to true if not set
func (e *MqExchange) IsEncrypted() bool {
	if e.Encrypted == nil {
		return true
	}
	return *e.Encrypted
}

// mqExchangeRequest is the request body for creating/updating an exchange (exchangeId is in URL, not body)
type mqExchangeRequest struct {
	Fifo      bool `json:"fifo,omitempty"`
	Encrypted bool `json:"encrypted"`
}

// MqRoutingRule represents a routing rule for exchange bindings
type MqRoutingRule struct {
	PropertyName string `json:"propertyName"`
	PropertyType string `json:"propertyType"`
	MatcherType  string `json:"matcherType"`
	Value        any    `json:"value"` // Can be string or []string depending on matcherType
}

// MqBinding represents a binding between an exchange and a queue
type MqBinding struct {
	QueueID      string          `json:"queueId"`
	ExchangeID   string          `json:"exchangeId,omitempty"`
	Fifo         bool            `json:"fifo"`
	RoutingRules []MqRoutingRule `json:"routingRules,omitempty"`
}

// MqDestination represents a destination (queue or exchange) in the API response
type MqDestination struct {
	Type                 string `json:"type"` // "queue" or "exchange"
	QueueID              string `json:"queueId,omitempty"`
	ExchangeID           string `json:"exchangeId,omitempty"`
	Fifo                 bool   `json:"fifo"`
	Encrypted            bool   `json:"encrypted"`
	MaxDeliveries        int    `json:"maxDeliveries,omitempty"`
	DeadLetterQueueID    string `json:"deadLetterQueueId,omitempty"`
	IsFallback           bool   `json:"isFallback,omitempty"`
	DefaultTtl           int64  `json:"defaultTtl,omitempty"`
	DefaultLockTtl       int64  `json:"defaultLockTtl,omitempty"`
	DefaultDeliveryDelay int64  `json:"defaultDeliveryDelay,omitempty"`
}

// GetMqDestinations retrieves all MQ destinations (queues and exchanges) for a region
func (client *AnypointClient) GetMqDestinations(orgID, envID, region string) ([]MqDestination, error) {
	reqPath := fmt.Sprintf("mq/admin/api/v1/organizations/%s/environments/%s/regions/%s/destinations", orgID, envID, region)
	req, err := client.newRequest("GET", reqPath, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}
	req.Header.Set("Accept", "application/json")

	res, err := client.HTTPClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to call Anypoint Platform")
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(res.Body)
		return nil, errors.Errorf("call to Anypoint Platform returned %d: %s", res.StatusCode, string(bodyBytes))
	}

	var destinations []MqDestination
	err = decodeResponseBody(res.Body, &destinations)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode response")
	}

	return destinations, nil
}

// GetMqQueue retrieves a specific queue by ID
func (client *AnypointClient) GetMqQueue(orgID, envID, region, queueID string) (*MqDestination, error) {
	reqPath := fmt.Sprintf("mq/admin/api/v1/organizations/%s/environments/%s/regions/%s/destinations/queues/%s", orgID, envID, region, queueID)
	req, err := client.newRequest("GET", reqPath, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}
	req.Header.Set("Accept", "application/json")

	res, err := client.HTTPClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to call Anypoint Platform")
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if res.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(res.Body)
		return nil, errors.Errorf("call to Anypoint Platform returned %d: %s", res.StatusCode, string(bodyBytes))
	}

	var destination MqDestination
	err = decodeResponseBody(res.Body, &destination)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode response")
	}

	return &destination, nil
}

// CreateMqQueue creates a new MQ queue
func (client *AnypointClient) CreateMqQueue(orgID, envID, region string, queue MqQueue) error {
	reqPath := fmt.Sprintf("mq/admin/api/v1/organizations/%s/environments/%s/regions/%s/destinations/queues/%s", orgID, envID, region, queue.QueueID)

	// Build request body without queueId (it's in the URL)
	reqBody := mqQueueCreateRequest{
		Type:                 queue.Type,
		Fifo:                 queue.Fifo,
		Encrypted:            queue.IsEncrypted(),
		MaxDeliveries:        queue.MaxDeliveries,
		DeadLetterQueueID:    queue.DeadLetterQueueID,
		DefaultTtl:           queue.DefaultTtl,
		DefaultLockTtl:       queue.DefaultLockTtl,
		DefaultDeliveryDelay: queue.DefaultDeliveryDelay,
	}

	buffer := new(bytes.Buffer)
	err := json.NewEncoder(buffer).Encode(reqBody)
	if err != nil {
		return errors.Wrap(err, "failed to encode request")
	}

	req, err := client.newRequest("PUT", reqPath, buffer)
	if err != nil {
		return errors.Wrap(err, "failed to create request")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	res, err := client.HTTPClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to call Anypoint Platform")
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(res.Body)
		return errors.Errorf("call to Anypoint Platform returned %d: %s", res.StatusCode, string(bodyBytes))
	}

	return nil
}

// UpdateMqQueue updates an existing MQ queue using PATCH
func (client *AnypointClient) UpdateMqQueue(orgID, envID, region string, queue MqQueue) error {
	reqPath := fmt.Sprintf("mq/admin/api/v1/organizations/%s/environments/%s/regions/%s/destinations/queues/%s", orgID, envID, region, queue.QueueID)

	// Build request body for update (PATCH)
	reqBody := mqQueueUpdateRequest{
		Type:                 "queue",
		Fifo:                 queue.Fifo,
		Encrypted:            queue.IsEncrypted(),
		MaxDeliveries:        queue.MaxDeliveries,
		DeadLetterQueueID:    queue.DeadLetterQueueID,
		IsFallback:           queue.IsFallback,
		DefaultTtl:           queue.DefaultTtl,
		DefaultLockTtl:       queue.DefaultLockTtl,
		DefaultDeliveryDelay: queue.DefaultDeliveryDelay,
	}

	buffer := new(bytes.Buffer)
	err := json.NewEncoder(buffer).Encode(reqBody)
	if err != nil {
		return errors.Wrap(err, "failed to encode request")
	}

	req, err := client.newRequest("PATCH", reqPath, buffer)
	if err != nil {
		return errors.Wrap(err, "failed to create request")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	res, err := client.HTTPClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to call Anypoint Platform")
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(res.Body)
		return errors.Errorf("call to Anypoint Platform returned %d: %s", res.StatusCode, string(bodyBytes))
	}

	return nil
}

// GetMqExchange retrieves a specific exchange by ID
func (client *AnypointClient) GetMqExchange(orgID, envID, region, exchangeID string) (*MqDestination, error) {
	reqPath := fmt.Sprintf("mq/admin/api/v1/organizations/%s/environments/%s/regions/%s/destinations/exchanges/%s", orgID, envID, region, exchangeID)
	req, err := client.newRequest("GET", reqPath, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}
	req.Header.Set("Accept", "application/json")

	res, err := client.HTTPClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to call Anypoint Platform")
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if res.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(res.Body)
		return nil, errors.Errorf("call to Anypoint Platform returned %d: %s", res.StatusCode, string(bodyBytes))
	}

	var destination MqDestination
	err = decodeResponseBody(res.Body, &destination)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode response")
	}

	return &destination, nil
}

// CreateMqExchange creates a new MQ exchange
func (client *AnypointClient) CreateMqExchange(orgID, envID, region string, exchange MqExchange) error {
	reqPath := fmt.Sprintf("mq/admin/api/v1/organizations/%s/environments/%s/regions/%s/destinations/exchanges/%s", orgID, envID, region, exchange.ExchangeID)

	// Build request body without exchangeId (it's in the URL)
	reqBody := mqExchangeRequest{
		Fifo:      exchange.Fifo,
		Encrypted: exchange.IsEncrypted(),
	}

	buffer := new(bytes.Buffer)
	err := json.NewEncoder(buffer).Encode(reqBody)
	if err != nil {
		return errors.Wrap(err, "failed to encode request")
	}

	req, err := client.newRequest("PUT", reqPath, buffer)
	if err != nil {
		return errors.Wrap(err, "failed to create request")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	res, err := client.HTTPClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to call Anypoint Platform")
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(res.Body)
		return errors.Errorf("call to Anypoint Platform returned %d: %s", res.StatusCode, string(bodyBytes))
	}

	return nil
}

// GetMqExchangeBindings retrieves all bindings for an exchange (including routing rules)
func (client *AnypointClient) GetMqExchangeBindings(orgID, envID, region, exchangeID string) ([]MqBinding, error) {
	reqPath := fmt.Sprintf("mq/admin/api/v1/organizations/%s/environments/%s/regions/%s/bindings/exchanges/%s?inclusion=ALL", orgID, envID, region, exchangeID)
	req, err := client.newRequest("GET", reqPath, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}
	req.Header.Set("Accept", "application/json")

	res, err := client.HTTPClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to call Anypoint Platform")
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if res.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(res.Body)
		return nil, errors.Errorf("call to Anypoint Platform returned %d: %s", res.StatusCode, string(bodyBytes))
	}

	var bindings []MqBinding
	err = decodeResponseBody(res.Body, &bindings)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode response")
	}

	return bindings, nil
}

// CreateMqBinding creates a binding between an exchange and a queue (without routing rules)
func (client *AnypointClient) CreateMqBinding(orgID, envID, region, exchangeID, queueID string) error {
	reqPath := fmt.Sprintf("mq/admin/api/v1/organizations/%s/environments/%s/regions/%s/bindings/exchanges/%s/queues/%s", orgID, envID, region, exchangeID, queueID)

	req, err := client.newRequest("PUT", reqPath, nil)
	if err != nil {
		return errors.Wrap(err, "failed to create request")
	}
	req.Header.Set("Accept", "application/json")

	res, err := client.HTTPClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to call Anypoint Platform")
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusNoContent {
		bodyBytes, _ := io.ReadAll(res.Body)
		return errors.Errorf("call to Anypoint Platform returned %d: %s", res.StatusCode, string(bodyBytes))
	}

	return nil
}

// MqRoutingRulesRequest represents the request body for updating routing rules
type MqRoutingRulesRequest struct {
	RoutingRules []MqRoutingRule `json:"routingRules"`
}

// UpdateMqBindingRoutingRules updates the routing rules for a binding
func (client *AnypointClient) UpdateMqBindingRoutingRules(orgID, envID, region, exchangeID, queueID string, routingRules []MqRoutingRule) error {
	reqPath := fmt.Sprintf("mq/admin/api/v1/organizations/%s/environments/%s/regions/%s/bindings/exchanges/%s/queues/%s/rules/routing", orgID, envID, region, exchangeID, queueID)

	requestBody := MqRoutingRulesRequest{RoutingRules: routingRules}
	buffer := new(bytes.Buffer)
	err := json.NewEncoder(buffer).Encode(requestBody)
	if err != nil {
		return errors.Wrap(err, "failed to encode request")
	}

	req, err := client.newRequest("PUT", reqPath, buffer)
	if err != nil {
		return errors.Wrap(err, "failed to create request")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	res, err := client.HTTPClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to call Anypoint Platform")
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusNoContent {
		bodyBytes, _ := io.ReadAll(res.Body)
		return errors.Errorf("call to Anypoint Platform returned %d: %s", res.StatusCode, string(bodyBytes))
	}

	return nil
}

// DeleteMqBinding deletes a binding between an exchange and a queue
func (client *AnypointClient) DeleteMqBinding(orgID, envID, region, exchangeID, queueID string) error {
	reqPath := fmt.Sprintf("mq/admin/api/v1/organizations/%s/environments/%s/regions/%s/bindings/exchanges/%s/queues/%s", orgID, envID, region, exchangeID, queueID)

	req, err := client.newRequest("DELETE", reqPath, nil)
	if err != nil {
		return errors.Wrap(err, "failed to create request")
	}

	res, err := client.HTTPClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to call Anypoint Platform")
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusNoContent && res.StatusCode != http.StatusNotFound {
		bodyBytes, _ := io.ReadAll(res.Body)
		return errors.Errorf("call to Anypoint Platform returned %d: %s", res.StatusCode, string(bodyBytes))
	}

	return nil
}
