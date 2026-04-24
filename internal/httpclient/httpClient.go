// Package httpclient contains an http client
package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"

	"github.com/google/go-querystring/query"
)

// Client manages making HTTP requests to the API.
type Client struct {
	baseURL    string
	apiKey     string
	userAgent  string
	httpClient *http.Client
	mu         sync.RWMutex // Protects token
	authToken  *string
}

// New creates a new internal HTTP client.
func New(baseURL, apiKey, userAgent string) *Client {
	return &Client{
		baseURL:    baseURL,
		apiKey:     apiKey,
		userAgent:  userAgent,
		httpClient: &http.Client{}, // Use default client, customize if needed (timeout, transport)
	}
}

// SetBaseURL updates the base URL used for requests.
func (c *Client) SetBaseURL(baseURL string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.baseURL = baseURL
}

// SetAuthToken updates the authentication token.
func (c *Client) SetAuthToken(token *string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.authToken = token
}

// Get makes a GET request.
func (c *Client) Get(ctx context.Context, path string, params any, target any) error {
	return c.doRequest(ctx, http.MethodGet, path, params, nil, target)
}

// Post makes a POST request.
func (c *Client) Post(ctx context.Context, path string, body any, target any) error {
	return c.doRequest(ctx, http.MethodPost, path, nil, body, target)
}

// Delete makes a DELETE request.
func (c *Client) Delete(ctx context.Context, path string, target any) error {
	return c.doRequest(ctx, http.MethodDelete, path, nil, nil, target)
}

// doRequest performs the actual HTTP request.
func (c *Client) doRequest(ctx context.Context, method, path string, params any, body any, target any) error {
	c.mu.RLock()
	currentBaseURL := c.baseURL
	currentToken := c.authToken
	c.mu.RUnlock()

	fullURL, err := url.Parse(currentBaseURL)
	if err != nil {
		return fmt.Errorf("invalid base URL: %w", err)
	}
	fullURL.Path += path // Assumes baseURL doesn't end with / and path starts with /

	// Encode query parameters if provided
	if params != nil {
		var v url.Values
		v, err = query.Values(params)
		if err != nil {
			return fmt.Errorf("failed to encode query parameters: %w", err)
		}
		fullURL.RawQuery = v.Encode()
	}

	// Encode request body if provided
	var reqBody io.Reader
	var contentType string
	if body != nil {
		var jsonData []byte
		jsonData, err = json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
		contentType = "application/json"
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL.String(), reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}
	req.Header.Set("Accept", "application/json")

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	// Add Authorization header if token exists
	if currentToken != nil && *currentToken != "" {
		req.Header.Set("Authorization", "Bearer "+*currentToken)
	}

	// Make the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Check status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("api request failed: status %d, body: %s", resp.StatusCode, string(respBodyBytes))
	}

	// Decode successful response if target is provided
	if target != nil {
		if err := json.Unmarshal(respBodyBytes, target); err != nil {
			return fmt.Errorf("failed to unmarshal response body: %w", err)
		}
	}

	return nil
}
