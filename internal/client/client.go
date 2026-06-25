package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
)

// Client is an authenticated HTTP client for the Nightlight Cloud API.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new Client and authenticates with the given credentials.
func NewClient(endpoint, username, password string) (*Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	c := &Client{
		baseURL:    endpoint,
		httpClient: &http.Client{Jar: jar},
	}
	if err := c.login(username, password); err != nil {
		return nil, fmt.Errorf("nightlight authentication failed: %w", err)
	}
	return c, nil
}

func (c *Client) login(username, password string) error {
	body, _ := json.Marshal(map[string]string{"username": username, "password": password})
	resp, err := c.httpClient.Post(c.baseURL+"/api/v1/auth/login", "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("status %d: %s", resp.StatusCode, string(b))
	}
	return nil
}

// request performs an authenticated API call. payload is JSON-encoded if non-nil; out is JSON-decoded if non-nil.
func (c *Client) request(method, path string, payload interface{}, out interface{}) error {
	var body io.Reader
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		body = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, c.baseURL+path, body)
	if err != nil {
		return err
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusNotFound {
		return ErrNotFound
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(b))
	}
	if out != nil && len(b) > 0 {
		return json.Unmarshal(b, out)
	}
	return nil
}

// ErrNotFound is returned when the API responds with 404.
var ErrNotFound = fmt.Errorf("resource not found")
