package arrclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client is a generic *Arr API client.
type Client struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

// NewClient creates a new *Arr API client.
func NewClient(baseURL, apiKey string, timeout time.Duration, sslVerify bool) *Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if !sslVerify {
		transport.TLSClientConfig.InsecureSkipVerify = true
	}
	return &Client{
		BaseURL: strings.TrimRight(baseURL, "/"),
		APIKey:  apiKey,
		HTTPClient: &http.Client{
			Timeout:   timeout,
			Transport: transport,
		},
	}
}

func (c *Client) doRequest(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	reqURL := c.BaseURL + path
	req, err := http.NewRequestWithContext(ctx, method, reqURL, body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("X-Api-Key", c.APIKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		errBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("api error %d: %s", resp.StatusCode, string(errBody))
	}
	return resp, nil
}

func (c *Client) get(ctx context.Context, path string, result any) error {
	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(result)
}

func (c *Client) post(ctx context.Context, path string, body io.Reader, result any) error {
	resp, err := c.doRequest(ctx, http.MethodPost, path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if result != nil {
		return json.NewDecoder(resp.Body).Decode(result)
	}
	return nil
}

// SystemStatus represents the status response from any *Arr app.
type SystemStatus struct {
	AppName string `json:"appName"`
	Version string `json:"version"`
}

// TestConnection verifies connectivity and returns system status.
func (c *Client) TestConnection(ctx context.Context, apiVersion string) (*SystemStatus, error) {
	var status SystemStatus
	if err := c.get(ctx, "/api/"+apiVersion+"/system/status", &status); err != nil {
		return nil, fmt.Errorf("test connection: %w", err)
	}
	return &status, nil
}

// CommandResponse is the response from a command endpoint.
type CommandResponse struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

// QueueRecord represents an item in the download queue.
type QueueRecord struct {
	ID     int    `json:"id"`
	Status string `json:"status"`
	Title  string `json:"title"`
	Size   int64  `json:"size"`
}

// QueueResponse wraps a paginated queue response.
type QueueResponse struct {
	TotalRecords int           `json:"totalRecords"`
	Records      []QueueRecord `json:"records"`
}

// IsPrivateIP checks if a URL points to a private/loopback address.
// Used for SSRF protection in test-connection.
func IsPrivateIP(rawURL string) (bool, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false, fmt.Errorf("parse url: %w", err)
	}
	host := parsed.Hostname()
	ip := net.ParseIP(host)
	if ip == nil {
		// Resolve hostname
		addrs, err := net.LookupHost(host)
		if err != nil {
			return false, fmt.Errorf("resolve host: %w", err)
		}
		for _, addr := range addrs {
			resolved := net.ParseIP(addr)
			if resolved != nil && (resolved.IsLoopback() || resolved.IsPrivate() || resolved.IsLinkLocalUnicast()) {
				return true, nil
			}
		}
		return false, nil
	}
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast(), nil
}
