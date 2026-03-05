package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// APIResponse matches fleet-api response shape.
type APIResponse struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data,omitempty"`
	Error    string          `json:"error,omitempty"`
	Duration int64           `json:"duration_ms,omitempty"`
}

// Client talks to fleet-api with API key auth.
type Client struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

// NewClient builds a client. apiKey is required for authenticated calls.
func NewClient(apiKey string) *Client {
	baseURL := os.Getenv("GRADIENT_API_URL")
	if baseURL == "" {
		baseURL = "https://fleet.usegradient.dev"
	}
	baseURL = strings.TrimSuffix(baseURL, "/")
	return &Client{
		BaseURL:    baseURL,
		APIKey:     apiKey,
		HTTPClient: &http.Client{},
	}
}

// Do performs a request and returns the parsed APIResponse and any error.
func (c *Client) Do(method, path string, body interface{}) (*APIResponse, error) {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(b)
	}
	u, err := url.JoinPath(c.BaseURL, path)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(method, u, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var out APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	if resp.StatusCode >= 400 {
		msg := out.Error
		if msg == "" {
			msg = resp.Status
		}
		return &out, fmt.Errorf("%s: %s", resp.Status, msg)
	}
	if !out.OK && out.Error != "" {
		return &out, fmt.Errorf("%s", out.Error)
	}
	return &out, nil
}

// Get is shorthand for Do("GET", path, nil).
func (c *Client) Get(path string) (*APIResponse, error) {
	return c.Do(http.MethodGet, path, nil)
}

// Post is shorthand for Do("POST", path, body).
func (c *Client) Post(path string, body interface{}) (*APIResponse, error) {
	return c.Do(http.MethodPost, path, body)
}

// Put is shorthand for Do("PUT", path, body).
func (c *Client) Put(path string, body interface{}) (*APIResponse, error) {
	return c.Do(http.MethodPut, path, body)
}

// Delete is shorthand for Do("DELETE", path, nil).
func (c *Client) Delete(path string) (*APIResponse, error) {
	return c.Do(http.MethodDelete, path, nil)
}

// DataInto unmarshals resp.Data into v. Use when resp.OK and resp.Data is JSON.
func DataInto(resp *APIResponse, v interface{}) error {
	if resp == nil || len(resp.Data) == 0 {
		return nil
	}
	return json.Unmarshal(resp.Data, v)
}
