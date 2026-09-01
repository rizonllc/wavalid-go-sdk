// Package wavalid is the Go SDK for the wavalid WhatsApp number validation API.
package wavalid

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type NumberStatus string

const (
	StatusValid   NumberStatus = "valid"
	StatusInvalid NumberStatus = "invalid"
	StatusLimit   NumberStatus = "limit"
)

type ValidateResult struct {
	PhoneNumber      string       `json:"phoneNumber"`
	Status           NumberStatus `json:"status"`
	CreditsRemaining int          `json:"creditsRemaining"`
}

type BulkResultItem struct {
	PhoneNumber string       `json:"phoneNumber"`
	Status      NumberStatus `json:"status"`
}

type ValidateBulkResult struct {
	Results          []BulkResultItem `json:"results"`
	CreditsUsed      int               `json:"creditsUsed"`
	CreditsRemaining int               `json:"creditsRemaining"`
}

// ApiError is returned for any non-2xx response from the wavalid API.
type ApiError struct {
	StatusCode int
	Message    string
	Code       string
}

func (e *ApiError) Error() string {
	return fmt.Sprintf("wavalid: %s (status %d)", e.Message, e.StatusCode)
}

type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a wavalid API client. baseURL is the root domain only
// (e.g. "https://wavalid.com") — do not append "/api" or "/api/v1".
func NewClient(apiKey, baseURL string) *Client {
	return &Client{
		apiKey:     apiKey,
		baseURL:    strings.TrimSuffix(baseURL, "/"),
		httpClient: http.DefaultClient,
	}
}

type validateInput struct {
	PhoneNumber string `json:"phoneNumber"`
	BatchID     *int   `json:"batchId,omitempty"`
}

// Validate checks a single phone number and deducts one credit if the check completes.
func (c *Client) Validate(ctx context.Context, phoneNumber string, batchID *int) (*ValidateResult, error) {
	var result ValidateResult
	if err := c.request(ctx, "/v1/validate", validateInput{PhoneNumber: phoneNumber, BatchID: batchID}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

type validateBulkInput struct {
	PhoneNumbers []string `json:"phoneNumbers"`
	BatchID      *int     `json:"batchId,omitempty"`
}

// ValidateBulk checks up to 100 numbers in one request and deducts one credit per number checked.
func (c *Client) ValidateBulk(ctx context.Context, phoneNumbers []string, batchID *int) (*ValidateBulkResult, error) {
	var result ValidateBulkResult
	if err := c.request(ctx, "/v1/validate/bulk", validateBulkInput{PhoneNumbers: phoneNumbers, BatchID: batchID}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) request(ctx context.Context, path string, body, out interface{}) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api"+path, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errPayload struct {
			StatusCode int    `json:"statusCode"`
			Message    string `json:"message"`
			Code       string `json:"code"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&errPayload)

		message := errPayload.Message
		if message == "" {
			message = "Something went wrong"
		}
		statusCode := errPayload.StatusCode
		if statusCode == 0 {
			statusCode = resp.StatusCode
		}
		return &ApiError{StatusCode: statusCode, Message: message, Code: errPayload.Code}
	}

	return json.NewDecoder(resp.Body).Decode(out)
}
