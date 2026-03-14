package deluge

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Client is a Deluge JSON-RPC client (Deluge WebUI API).
type Client struct {
	BaseURL    string
	Password   string
	HTTPClient *http.Client

	mu            sync.Mutex
	authenticated bool
	requestID     atomic.Int64
}

// NewClient creates a new Deluge JSON-RPC client.
func NewClient(baseURL, password string, timeout time.Duration) *Client {
	jar, _ := cookiejar.New(nil)
	return &Client{
		BaseURL:  strings.TrimRight(baseURL, "/"),
		Password: password,
		HTTPClient: &http.Client{
			Timeout: timeout,
			Jar:     jar,
		},
	}
}

// jsonRPCRequest is a Deluge JSON-RPC request.
type jsonRPCRequest struct {
	ID     int64       `json:"id"`
	Method string      `json:"method"`
	Params interface{} `json:"params"`
}

// jsonRPCResponse is a Deluge JSON-RPC response.
type jsonRPCResponse struct {
	ID     int64           `json:"id"`
	Result json.RawMessage `json:"result"`
	Error  *jsonRPCError   `json:"error"`
}

type jsonRPCError struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// login authenticates with the Deluge Web UI and ensures the daemon is connected.
func (c *Client) login(ctx context.Context) error {
	result, err := c.call(ctx, "auth.login", []interface{}{c.Password})
	if err != nil {
		return fmt.Errorf("auth.login: %w", err)
	}
	var ok bool
	if err := json.Unmarshal(result, &ok); err != nil {
		return fmt.Errorf("decode login result: %w", err)
	}
	if !ok {
		return fmt.Errorf("login rejected: invalid password")
	}
	c.authenticated = true

	// Ensure the WebUI is connected to its daemon.
	if err := c.ensureDaemonConnected(ctx); err != nil {
		return fmt.Errorf("daemon connect: %w", err)
	}
	return nil
}

// ensureDaemonConnected checks if the WebUI has an active daemon connection,
// and if not, connects to the first configured host.
func (c *Client) ensureDaemonConnected(ctx context.Context) error {
	raw, err := c.call(ctx, "web.connected", nil)
	if err != nil {
		return err
	}
	var connected bool
	if err := json.Unmarshal(raw, &connected); err == nil && connected {
		return nil
	}

	// Get configured hosts and connect to the first one.
	raw, err = c.call(ctx, "web.get_hosts", nil)
	if err != nil {
		return fmt.Errorf("web.get_hosts: %w", err)
	}
	var hosts []json.RawMessage
	if err := json.Unmarshal(raw, &hosts); err != nil || len(hosts) == 0 {
		return fmt.Errorf("no daemon hosts configured")
	}
	// Each host is [id, ip, port, user]; extract the id (first element).
	var firstHost []json.RawMessage
	if err := json.Unmarshal(hosts[0], &firstHost); err != nil || len(firstHost) == 0 {
		return fmt.Errorf("invalid host entry")
	}
	var hostID string
	if err := json.Unmarshal(firstHost[0], &hostID); err != nil {
		return fmt.Errorf("decode host id: %w", err)
	}
	_, err = c.call(ctx, "web.connect", []interface{}{hostID})
	return err
}

// ensureAuth logs in if not already authenticated.
func (c *Client) ensureAuth(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.authenticated {
		return nil
	}
	return c.login(ctx)
}

// call performs a raw JSON-RPC call (no auth check).
func (c *Client) call(ctx context.Context, method string, params interface{}) (json.RawMessage, error) {
	if params == nil {
		params = []interface{}{}
	}
	reqID := c.requestID.Add(1)
	payload := jsonRPCRequest{
		ID:     reqID,
		Method: method,
		Params: params,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}
	reqURL := c.BaseURL + "/json"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
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

// authenticatedCall performs a JSON-RPC call, re-authenticating on 401/rejection.
func (c *Client) authenticatedCall(ctx context.Context, method string, params interface{}) (json.RawMessage, error) {
	if err := c.ensureAuth(ctx); err != nil {
		return nil, err
	}
	result, err := c.call(ctx, method, params)
	if err != nil {
		// On auth error, try re-login once.
		c.mu.Lock()
		c.authenticated = false
		c.mu.Unlock()
		if loginErr := c.ensureAuth(ctx); loginErr != nil {
			return nil, err // return original error
		}
		return c.call(ctx, method, params)
	}
	return result, nil
}

// Torrent represents a Deluge torrent.
type Torrent struct {
	Hash          string  `json:"hash"`
	Name          string  `json:"name"`
	State         string  `json:"state"`
	TotalSize     int64   `json:"total_size"`
	Progress      float64 `json:"progress"`
	DownloadSpeed float64 `json:"download_payload_rate"`
	UploadSpeed   float64 `json:"upload_payload_rate"`
	Ratio         float64 `json:"ratio"`
	ETA           int64   `json:"eta"`
	NumSeeds      int     `json:"num_seeds"`
	NumPeers      int     `json:"num_peers"`
	SavePath      string  `json:"save_path"`
	TimeAdded     float64 `json:"time_added"`
	SeedingTime   int64   `json:"seeding_time"`
	TotalDone     int64   `json:"total_done"`
	TrackerHost   string  `json:"tracker_host"`
	Label         string  `json:"label"`
}

var defaultFields = []string{
	"hash", "name", "state", "total_size", "progress",
	"download_payload_rate", "upload_payload_rate", "ratio", "eta",
	"num_seeds", "num_peers", "save_path", "time_added", "seeding_time",
	"total_done", "tracker_host", "label",
}

// GetTorrents returns all torrents with default fields.
func (c *Client) GetTorrents(ctx context.Context) ([]Torrent, error) {
	params := []interface{}{
		map[string]interface{}{}, // filter_dict — empty = all
		defaultFields,
	}
	raw, err := c.authenticatedCall(ctx, "core.get_torrents_status", params)
	if err != nil {
		return nil, err
	}
	// Deluge returns a map of hash -> torrent fields.
	var torrentMap map[string]Torrent
	if err := json.Unmarshal(raw, &torrentMap); err != nil {
		return nil, fmt.Errorf("decode torrents: %w", err)
	}
	torrents := make([]Torrent, 0, len(torrentMap))
	for hash, t := range torrentMap {
		t.Hash = hash
		torrents = append(torrents, t)
	}
	return torrents, nil
}

// PauseTorrents pauses the specified torrents by hash.
func (c *Client) PauseTorrents(ctx context.Context, hashes []string) error {
	_, err := c.authenticatedCall(ctx, "core.pause_torrents", []interface{}{hashes})
	return err
}

// ResumeTorrents resumes the specified torrents by hash.
func (c *Client) ResumeTorrents(ctx context.Context, hashes []string) error {
	_, err := c.authenticatedCall(ctx, "core.resume_torrents", []interface{}{hashes})
	return err
}

// RecheckTorrents triggers a data integrity recheck for the specified torrents.
func (c *Client) RecheckTorrents(ctx context.Context, hashes []string) error {
	_, err := c.authenticatedCall(ctx, "core.force_recheck", []interface{}{hashes})
	return err
}

// DeleteTorrents removes the specified torrents. If removeData is true, downloaded data is also removed.
func (c *Client) DeleteTorrents(ctx context.Context, hashes []string, removeData bool) error {
	for _, hash := range hashes {
		_, err := c.authenticatedCall(ctx, "core.remove_torrent", []interface{}{hash, removeData})
		if err != nil {
			return fmt.Errorf("remove %s: %w", hash, err)
		}
	}
	return nil
}

// GetVersion returns the Deluge daemon version.
func (c *Client) GetVersion(ctx context.Context) (string, error) {
	raw, err := c.authenticatedCall(ctx, "daemon.get_version", nil)
	if err != nil {
		return "", err
	}
	var version string
	if err := json.Unmarshal(raw, &version); err != nil {
		return "", fmt.Errorf("decode version: %w", err)
	}
	return version, nil
}

// TestConnection verifies Deluge is reachable and credentials are valid.
func (c *Client) TestConnection(ctx context.Context) (string, error) {
	return c.GetVersion(ctx)
}

// AddTorrentByURL adds a torrent via magnet URI or HTTP URL.
func (c *Client) AddTorrentByURL(ctx context.Context, torrentURL string, options map[string]interface{}) error {
	if options == nil {
		options = map[string]interface{}{}
	}
	_, err := c.authenticatedCall(ctx, "core.add_torrent_url", []interface{}{torrentURL, options})
	return err
}
