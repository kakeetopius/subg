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
	headers    map[string]string
	httpClient *http.Client
	mu         sync.RWMutex // Protects headers and baseURL
}

// New creates a new internal HTTP client.
func New() *Client {
	return &Client{
		httpClient: &http.Client{},
		headers:    make(map[string]string),
	}
}

func (c *Client) WithAPIKey(key string) *Client {
	return c.WithHeader("Api-Key", key)
}

func (c *Client) WithUserAgent(userAgent string) *Client {
	return c.WithHeader("User-Agent", userAgent)
}

func (c *Client) WithCookie(cookie string) *Client {
	return c.WithHeader("Cookie", cookie)
}

// WithBaseURL updates the base URL used for requests.
func (c *Client) WithBaseURL(baseURL string) *Client {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.baseURL = baseURL

	return c
}

func (c *Client) WithHeader(key string, value string) *Client {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.headers == nil {
		c.headers = make(map[string]string)
	}
	c.headers[key] = value

	return c
}

// Get makes a GET request.
func (c *Client) Get(ctx context.Context, path string, params any) (*http.Response, error) {
	return c.doRequest(ctx, http.MethodGet, path, params, nil)
}

// GetJSON makes a GET request and intialises target with JSON response if not nil
func (c *Client) GetJSON(ctx context.Context, path string, params any, target any) error {
	return c.doRequestJSON(ctx, http.MethodGet, path, params, nil, target)
}

// Post makes a POST request.
func (c *Client) Post(ctx context.Context, path string, body any) (*http.Response, error) {
	return c.doRequest(ctx, http.MethodPost, path, nil, body)
}

// PostJSON makes a POST request and intialises target with JSON response if not nil
func (c *Client) PostJSON(ctx context.Context, path string, body any, target any) error {
	return c.doRequestJSON(ctx, http.MethodPost, path, nil, body, target)
}

// Delete makes a DELETE request.
func (c *Client) Delete(ctx context.Context, path string) (*http.Response, error) {
	return c.doRequest(ctx, http.MethodDelete, path, nil, nil)
}

// DeleteJSON makes a DELETE request and intialises target with JSON response if not nil
func (c *Client) DeleteJSON(ctx context.Context, path string, target any) error {
	return c.doRequestJSON(ctx, http.MethodDelete, path, nil, nil, target)
}

// doRequest performs the actual HTTP request.
func (c *Client) doRequest(ctx context.Context, method, path string, params any, body any) (*http.Response, error) {
	c.mu.RLock()
	currentBaseURL := c.baseURL
	headers := c.headers
	c.mu.RUnlock()

	fullURL, err := url.Parse(currentBaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}
	fullURL.Path += path // Assumes baseURL doesn't end with / and path starts with /

	// Encode query parameters if provided
	if params != nil {
		var v url.Values
		v, err = query.Values(params)
		if err != nil {
			return nil, fmt.Errorf("failed to encode query parameters: %w", err)
		}
		fullURL.RawQuery = v.Encode()
	}

	// Encode request body if provided
	var reqBody io.Reader
	if body != nil {
		var jsonData []byte
		jsonData, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
		headers["Content-Type"] = "application/json"
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL.String(), reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	for key, value := range headers {
		if value != "" {
			req.Header.Set(key, value)
		}
	}

	// Make the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	return resp, nil
}

// doRequest performs the actual HTTP request. If target is not nil, it assumes a response that is json is returned.
func (c *Client) doRequestJSON(ctx context.Context, method, path string, params any, body any, target any) error {
	c.WithHeader("Accept", "application/json")

	resp, err := c.doRequest(ctx, method, path, params, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Read response body
	respBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Check status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("api request failed: status %s, body: %s", resp.Status, string(respBodyBytes))
	}

	// Decode successful response if target is provided
	if target != nil {
		if err = json.Unmarshal(respBodyBytes, target); err != nil {
			return fmt.Errorf("failed to unmarshal response body: %w", err)
		}
	}

	return nil
}
