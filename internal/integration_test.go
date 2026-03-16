package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	baseURL       = "http://localhost:9705"
	adminUser     = "admin"
	adminPassword = "admin123"
)

var (
	client        = &http.Client{Timeout: 10 * time.Second}
	sessionCookie string
)

func init() {
	// Login once to get session cookie
	payload := map[string]string{
		"username": adminUser,
		"password": adminPassword,
	}
	body, _ := json.Marshal(payload)

	loginResp, err := client.Post(
		baseURL+"/api/auth/login",
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		return
	}
	defer loginResp.Body.Close()

	if loginResp.StatusCode != http.StatusOK {
		return
	}

	// Extract cookie from Set-Cookie headers
	for _, c := range loginResp.Cookies() {
		if c.Name == "lurkarr_session" {
			sessionCookie = c.Value
			break
		}
	}
}

// newRequest creates a request with auth cookie
func newRequest(method, path string, body []byte) *http.Request {
	var req *http.Request
	if body != nil {
		req, _ = http.NewRequest(method, baseURL+path, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, _ = http.NewRequest(method, baseURL+path, nil)
	}

	if sessionCookie != "" {
		req.Header.Set("Cookie", fmt.Sprintf("lurkarr_session=%s", sessionCookie))
	}
	return req
}

// TestLiveAppInstancesEndpoint verifies the live server returns all configured apps
func TestLiveAppInstancesEndpoint(t *testing.T) {
	req := newRequest("GET", "/api/instances", nil)
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var instances map[string][]map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&instances)

	// Verify we have at least the seeded apps
	assert.True(t, len(instances) > 0, "Should have configured app types")
	assert.NotNil(t, instances["sonarr"], "Should have Sonarr instances")
	assert.NotNil(t, instances["radarr"], "Should have Radarr instances")
}

// TestLiveHealthCheck verifies Sonarr health check works against live container
func TestLiveHealthCheck(t *testing.T) {
	// First get the instance IDs
	req := newRequest("GET", "/api/instances/sonarr", nil)
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var instances []map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&instances)

	require.NotEmpty(t, instances, "Should have Sonarr instances")
	sonarrID := instances[0]["id"].(string)

	// Now health check that instance
	healthReq := newRequest("GET", fmt.Sprintf("/api/instances/%s/health", sonarrID), nil)
	healthResp, err := client.Do(healthReq)
	require.NoError(t, err)
	defer healthResp.Body.Close()

	assert.Equal(t, http.StatusOK, healthResp.StatusCode, "Sonarr health check should succeed")

	var health map[string]interface{}
	json.NewDecoder(healthResp.Body).Decode(&health)

	assert.Equal(t, "ok", health["status"])
	assert.Equal(t, "Sonarr", health["app"])
	assert.NotNil(t, health["version"])
}

// TestAllAppsHealthy verifies all configured apps are responding
func TestAllAppsHealthy(t *testing.T) {
	// Get all instances
	req := newRequest("GET", "/api/instances", nil)
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var allInstances map[string][]map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&allInstances)

	// Test each one
	for appType, instances := range allInstances {
		for _, instance := range instances {
			instanceID := instance["id"].(string)
			name := instance["name"].(string)

			healthReq := newRequest("GET", fmt.Sprintf("/api/instances/%s/health", instanceID), nil)
			healthResp, err := client.Do(healthReq)
			assert.NoError(t, err, "Failed to health check %s: %s", appType, name)

			if healthResp != nil && healthResp.StatusCode == http.StatusOK {
				var health map[string]interface{}
				json.NewDecoder(healthResp.Body).Decode(&health)
				assert.Equal(t, "ok", health["status"],
					"App %s (%s) should be healthy", name, appType)
				healthResp.Body.Close()
			}
		}
	}
}

// TestActivityEndpointWorks verifies activity endpoint responds (even if empty)
func TestActivityEndpointWorks(t *testing.T) {
	req := newRequest("GET", "/api/activity", nil)
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Activity endpoint should respond")

	var activity map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&activity)
	_, hasItems := activity["items"]
	assert.True(t, hasItems, "Activity should have items field")
}

// TestStatisticsEndpointWorks verifies stats endpoint is accessible
func TestStatisticsEndpointWorks(t *testing.T) {
	req := newRequest("GET", "/api/stats", nil)
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Stats endpoint should respond")

	var stats interface{}
	json.NewDecoder(resp.Body).Decode(&stats)
	assert.NotNil(t, stats, "Stats should be returned")
}

// TestDownloadClientsEndpoint verifies download clients endpoint works
func TestDownloadClientsEndpoint(t *testing.T) {
	req := newRequest("GET", "/api/download-clients", nil)
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Download clients endpoint should be accessible")
}

// TestSchedulesEndpoint verifies schedules endpoint works
func TestSchedulesEndpoint(t *testing.T) {
	req := newRequest("GET", "/api/schedules", nil)
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Schedules endpoint should be accessible")
}

// TestHistoryEndpoint verifies history endpoint works
func TestHistoryEndpoint(t *testing.T) {
	req := newRequest("GET", "/api/history", nil)
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "History endpoint should be accessible")

	var history map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&history)

	items, hasItems := history["items"]
	assert.True(t, hasItems, "History should have items field")
	assert.NotNil(t, items, "Items should exist (may be empty)")
}

// TestUnauthorizedAccessBlocked verifies endpoints require auth
func TestUnauthorizedAccessBlocked(t *testing.T) {
	// Try to access protected endpoint without auth
	req, _ := http.NewRequest("GET", baseURL+"/api/instances", nil)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Should be either 401 or redirected to login
	assert.True(t, resp.StatusCode == http.StatusUnauthorized ||
		resp.StatusCode == http.StatusFound ||
		resp.StatusCode == http.StatusTemporaryRedirect,
		"Protected endpoint should require authentication, got: %d", resp.StatusCode)
}

// TestHealthEndpointPublic verifies health check is public (no auth required)
func TestHealthEndpointPublic(t *testing.T) {
	req, _ := http.NewRequest("GET", baseURL+"/healthz", nil)
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Health endpoint should be public")
}

// TestFrontendLoads verifies the SPA loads
func TestFrontendLoads(t *testing.T) {
	req, _ := http.NewRequest("GET", baseURL+"/", nil)
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Frontend should load")

	// Check it returns HTML
	contentType := resp.Header.Get("Content-Type")
	assert.Contains(t, contentType, "text/html", "Should serve HTML")
}
