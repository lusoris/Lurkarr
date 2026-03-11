package nzbget

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
)

// Client is an NZBGet JSON-RPC client.
type Client struct {
	BaseURL    string
	Username   string
	Password   string
	HTTPClient *http.Client

	requestID atomic.Int64
}

// NewClient creates a new NZBGet client.
func NewClient(baseURL, username, password string, timeout time.Duration) *Client {
	return &Client{
		BaseURL:  strings.TrimRight(baseURL, "/"),
		Username: username,
		Password: password,
		HTTPClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// jsonRPCRequest is an NZBGet JSON-RPC request.
type jsonRPCRequest struct {
	Method  string      `json:"method"`
	ID      int64       `json:"id"`
	Params  interface{} `json:"params,omitempty"`
	Version string      `json:"jsonrpc"`
}

// jsonRPCResponse is an NZBGet JSON-RPC response.
type jsonRPCResponse struct {
	ID     int64           `json:"id"`
	Result json.RawMessage `json:"result"`
	Error  *jsonRPCError   `json:"error"`
}

type jsonRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// call performs a JSON-RPC call to NZBGet.
func (c *Client) call(ctx context.Context, method string, params ...interface{}) (json.RawMessage, error) {
	reqID := c.requestID.Add(1)
	payload := jsonRPCRequest{
		Method:  method,
		ID:      reqID,
		Version: "2.0",
	}
	if len(params) > 0 {
		payload.Params = params
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}
	reqURL := c.BaseURL + "/jsonrpc"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.Username != "" {
		req.SetBasicAuth(c.Username, c.Password)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("authentication failed: check username/password")
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("http error %d: %s", resp.StatusCode, string(respBody))
	}

	var rpcResp jsonRPCResponse
	if err := json.Unmarshal(respBody, &rpcResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	if rpcResp.Error != nil {
		return nil, fmt.Errorf("rpc error %d: %s", rpcResp.Error.Code, rpcResp.Error.Message)
	}
	return rpcResp.Result, nil
}

// QueueItem represents an item in the NZBGet download queue.
type QueueItem struct {
	NZBID           int    `json:"NZBID"`
	NZBName         string `json:"NZBName"`
	NZBFilename     string `json:"NZBFilename"`
	Category        string `json:"Category"`
	FileSizeLo      int64  `json:"FileSizeLo"`
	FileSizeHi      int64  `json:"FileSizeHi"`
	RemainingSizeLo int64  `json:"RemainingSizeLo"`
	RemainingSizeHi int64  `json:"RemainingSizeHi"`
	Status          string `json:"Status"`
	TotalArticles   int    `json:"TotalArticles"`
	Health          int    `json:"Health"`
}

// HistoryItem represents a completed download in NZBGet.
type HistoryItem struct {
	NZBID        int    `json:"NZBID"`
	NZBName      string `json:"NZBName"`
	NZBFilename  string `json:"NZBFilename"`
	Category     string `json:"Category"`
	FileSizeLo   int64  `json:"FileSizeLo"`
	FileSizeHi   int64  `json:"FileSizeHi"`
	Status       string `json:"Status"`
	DestDir      string `json:"DestDir"`
	HealthStatus string `json:"HealthStatus,omitempty"`
}

// StatusInfo represents NZBGet server status.
type StatusInfo struct {
	RemainingSizeLo int64 `json:"RemainingSizeLo"`
	RemainingSizeHi int64 `json:"RemainingSizeHi"`
	DownloadRate    int64 `json:"DownloadRate"`
	DownloadPaused  bool  `json:"DownloadPaused"`
	ThreadCount     int   `json:"ThreadCount"`
	ServerStandBy   bool  `json:"ServerStandBy"`
	FreeDiskSpaceLo int64 `json:"FreeDiskSpaceLo"`
	FreeDiskSpaceHi int64 `json:"FreeDiskSpaceHi"`
	UpTimeSec       int64 `json:"UpTimeSec"`
}

// GetQueue returns the current download queue.
func (c *Client) GetQueue(ctx context.Context) ([]QueueItem, error) {
	raw, err := c.call(ctx, "listgroups")
	if err != nil {
		return nil, err
	}
	var items []QueueItem
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil, fmt.Errorf("decode queue: %w", err)
	}
	return items, nil
}

// GetHistory returns completed downloads.
func (c *Client) GetHistory(ctx context.Context) ([]HistoryItem, error) {
	raw, err := c.call(ctx, "history")
	if err != nil {
		return nil, err
	}
	var items []HistoryItem
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil, fmt.Errorf("decode history: %w", err)
	}
	return items, nil
}

// GetStatus returns NZBGet server status.
func (c *Client) GetStatus(ctx context.Context) (*StatusInfo, error) {
	raw, err := c.call(ctx, "status")
	if err != nil {
		return nil, err
	}
	var status StatusInfo
	if err := json.Unmarshal(raw, &status); err != nil {
		return nil, fmt.Errorf("decode status: %w", err)
	}
	return &status, nil
}

// Pause pauses the download queue.
func (c *Client) Pause(ctx context.Context) error {
	raw, err := c.call(ctx, "pausedownload")
	if err != nil {
		return err
	}
	var ok bool
	if err := json.Unmarshal(raw, &ok); err != nil {
		return fmt.Errorf("decode pause result: %w", err)
	}
	if !ok {
		return fmt.Errorf("pause failed")
	}
	return nil
}

// Resume resumes the download queue.
func (c *Client) Resume(ctx context.Context) error {
	raw, err := c.call(ctx, "resumedownload")
	if err != nil {
		return err
	}
	var ok bool
	if err := json.Unmarshal(raw, &ok); err != nil {
		return fmt.Errorf("decode resume result: %w", err)
	}
	if !ok {
		return fmt.Errorf("resume failed")
	}
	return nil
}

// DeleteItem removes an item from the queue or history by NZBID.
func (c *Client) DeleteItem(ctx context.Context, nzbID int) error {
	// editqueue: Command, Param, [IDs]
	raw, err := c.call(ctx, "editqueue", "GroupDelete", "", nzbID)
	if err != nil {
		return err
	}
	var ok bool
	if err := json.Unmarshal(raw, &ok); err != nil {
		return fmt.Errorf("decode delete result: %w", err)
	}
	if !ok {
		return fmt.Errorf("delete failed for NZBID %d", nzbID)
	}
	return nil
}

// GetVersion returns the NZBGet version string.
func (c *Client) GetVersion(ctx context.Context) (string, error) {
	raw, err := c.call(ctx, "version")
	if err != nil {
		return "", err
	}
	var version string
	if err := json.Unmarshal(raw, &version); err != nil {
		return "", fmt.Errorf("decode version: %w", err)
	}
	return version, nil
}

// TestConnection verifies NZBGet is reachable and credentials are valid.
func (c *Client) TestConnection(ctx context.Context) (string, error) {
	return c.GetVersion(ctx)
}
