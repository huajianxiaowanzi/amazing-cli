// Package codex provides functionality to fetch Codex token usage information.
package codex

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"
)

// RPCRateLimitWindow represents a rate limit window from Codex RPC.
type RPCRateLimitWindow struct {
	UsedPercent       float64 `json:"usedPercent"`
	WindowDurationMin int     `json:"windowDurationMins,omitempty"`
	ResetsAt          int64   `json:"resetsAt,omitempty"`
}

// RPCRateLimitSnapshot represents the full rate limit snapshot from Codex RPC.
type RPCRateLimitSnapshot struct {
	Primary   *RPCRateLimitWindow `json:"primary,omitempty"`
	Secondary *RPCRateLimitWindow `json:"secondary,omitempty"`
	Credits   *struct {
		HasCredits bool   `json:"hasCredits"`
		Unlimited  bool   `json:"unlimited"`
		Balance    string `json:"balance,omitempty"`
	} `json:"credits,omitempty"`
}

// RPCRateLimitsResponse is the response from account/rateLimits/read.
type RPCRateLimitsResponse struct {
	RateLimits RPCRateLimitSnapshot `json:"rateLimits"`
}

// RPCAccountResponse is the response from account/read.
type RPCAccountResponse struct {
	Account             *RPCAccountDetails `json:"account,omitempty"`
	RequiresOpenAIAuth  bool               `json:"requiresOpenaiAuth,omitempty"`
}

// RPCAccountDetails contains account details.
type RPCAccountDetails struct {
	Type     string `json:"type"`
	Email    string `json:"email,omitempty"`
	PlanType string `json:"planType,omitempty"`
}

// CodexRPCClient is a client for communicating with codex app-server via JSON-RPC.
type CodexRPCClient struct {
	cmd        *exec.Cmd
	stdin      io.WriteCloser
	stdout     *bufio.Scanner
	stderr     io.ReadCloser
	mu         sync.Mutex
	nextID     int
	lineChan   chan string
	errChan    chan error
	cancelFunc context.CancelFunc
}

// NewCodexRPCClient starts codex app-server and returns a client for RPC communication.
func NewCodexRPCClient(ctx context.Context) (*CodexRPCClient, error) {
	// Find codex binary
	codexPath, err := exec.LookPath("codex")
	if err != nil {
		return nil, fmt.Errorf("codex CLI not found: %w", err)
	}

	// Create context with cancel for cleanup
	ctx, cancel := context.WithCancel(ctx)

	// Start codex app-server with safe flags
	cmd := exec.CommandContext(ctx, codexPath, "-s", "read-only", "-a", "untrusted", "app-server")
	cmd.Env = os.Environ()

	stdin, err := cmd.StdinPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to start codex app-server: %w", err)
	}

	client := &CodexRPCClient{
		cmd:        cmd,
		stdin:      stdin,
		stdout:     bufio.NewScanner(stdout),
		stderr:     stderr,
		nextID:     1,
		lineChan:   make(chan string, 10),
		errChan:    make(chan error, 1),
		cancelFunc: cancel,
	}

	// Start reading stdout in background
	go client.readLines()

	return client, nil
}

// readLines reads lines from stdout in a goroutine.
func (c *CodexRPCClient) readLines() {
	for c.stdout.Scan() {
		c.lineChan <- c.stdout.Text()
	}
	if err := c.stdout.Err(); err != nil {
		select {
		case c.errChan <- err:
		default:
		}
	}
	close(c.lineChan)
}

// Close terminates the codex app-server process.
func (c *CodexRPCClient) Close() {
	c.cancelFunc()
	if c.stdin != nil {
		c.stdin.Close()
	}
	if c.cmd != nil && c.cmd.Process != nil {
		c.cmd.Process.Kill()
		c.cmd.Wait()
	}
}

// sendRequest sends a JSON-RPC request and waits for response.
func (c *CodexRPCClient) sendRequest(ctx context.Context, method string, params interface{}) (json.RawMessage, error) {
	c.mu.Lock()
	id := c.nextID
	c.nextID++
	c.mu.Unlock()

	// Build request
	request := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"method":  method,
	}
	if params != nil {
		request["params"] = params
	} else {
		request["params"] = map[string]interface{}{}
	}

	data, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Send request
	if _, err := c.stdin.Write(append(data, '\n')); err != nil {
		return nil, fmt.Errorf("failed to write request: %w", err)
	}

	// Wait for response with matching ID
	timeout := time.After(15 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timeout:
			return nil, fmt.Errorf("timeout waiting for response")
		case err := <-c.errChan:
			return nil, fmt.Errorf("error reading stdout: %w", err)
		case line, ok := <-c.lineChan:
			if !ok {
				return nil, fmt.Errorf("stdout closed")
			}

			var response struct {
				ID     interface{}     `json:"id"`
				Result json.RawMessage `json:"result,omitempty"`
				Error  *struct {
					Code    int    `json:"code"`
					Message string `json:"message"`
				} `json:"error,omitempty"`
			}

			if err := json.Unmarshal([]byte(line), &response); err != nil {
				// Not a valid JSON, might be a notification, skip
				continue
			}

			// Check if this is a notification (no ID)
			if response.ID == nil {
				continue
			}

			// Check if ID matches
			responseID := 0
			switch v := response.ID.(type) {
			case float64:
				responseID = int(v)
			case int:
				responseID = v
			}

			if responseID != id {
				continue
			}

			if response.Error != nil {
				return nil, fmt.Errorf("RPC error: %s", response.Error.Message)
			}

			return response.Result, nil
		}
	}
}

// sendNotification sends a JSON-RPC notification (no response expected).
func (c *CodexRPCClient) sendNotification(method string, params interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	request := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  method,
	}
	if params != nil {
		request["params"] = params
	} else {
		request["params"] = map[string]interface{}{}
	}

	data, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	if _, err := c.stdin.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("failed to write notification: %w", err)
	}

	return nil
}

// Initialize sends the initialize request to codex app-server.
func (c *CodexRPCClient) Initialize(ctx context.Context) error {
	params := map[string]interface{}{
		"clientInfo": map[string]interface{}{
			"name":    "amazing-cli",
			"version": "1.0.0",
		},
	}

	_, err := c.sendRequest(ctx, "initialize", params)
	if err != nil {
		return err
	}

	// Send initialized notification
	return c.sendNotification("initialized", nil)
}

// FetchRateLimits fetches the rate limits from codex app-server.
func (c *CodexRPCClient) FetchRateLimits(ctx context.Context) (*RPCRateLimitsResponse, error) {
	result, err := c.sendRequest(ctx, "account/rateLimits/read", nil)
	if err != nil {
		return nil, err
	}

	var response RPCRateLimitsResponse
	if err := json.Unmarshal(result, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal rate limits: %w", err)
	}

	return &response, nil
}

// FetchAccount fetches account information from codex app-server.
func (c *CodexRPCClient) FetchAccount(ctx context.Context) (*RPCAccountResponse, error) {
	result, err := c.sendRequest(ctx, "account/read", nil)
	if err != nil {
		return nil, err
	}

	var response RPCAccountResponse
	if err := json.Unmarshal(result, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal account: %w", err)
	}

	return &response, nil
}

// FetchUsageViaRPC fetches usage information using the RPC client.
func FetchUsageViaRPC(ctx context.Context) (UsageInfo, error) {
	client, err := NewCodexRPCClient(ctx)
	if err != nil {
		return UsageInfo{}, err
	}
	defer client.Close()

	// Initialize the connection
	if err := client.Initialize(ctx); err != nil {
		return UsageInfo{}, fmt.Errorf("failed to initialize: %w", err)
	}

	// Fetch rate limits
	rateLimits, err := client.FetchRateLimits(ctx)
	if err != nil {
		return UsageInfo{}, fmt.Errorf("failed to fetch rate limits: %w", err)
	}

	// Convert RPC response to UsageInfo
	return convertRPCToUsageInfo(rateLimits)
}

// convertRPCToUsageInfo converts RPC rate limits to UsageInfo.
func convertRPCToUsageInfo(resp *RPCRateLimitsResponse) (UsageInfo, error) {
	if resp.RateLimits.Primary == nil && resp.RateLimits.Secondary == nil {
		return UsageInfo{}, fmt.Errorf("no rate limit data available")
	}

	now := time.Now()
	
	// Parse primary (5h limit) - store remaining percentage
	var fiveHourInfo LimitInfo
	if resp.RateLimits.Primary != nil {
		used := int(resp.RateLimits.Primary.UsedPercent)
		remaining := 100 - used
		if remaining < 0 {
			remaining = 0
		}
		fiveHourInfo.Percentage = remaining // Store remaining, not used
		
		resetDesc := ""
		if resp.RateLimits.Primary.ResetsAt > 0 {
			resetTime := time.Unix(resp.RateLimits.Primary.ResetsAt, 0)
			resetDesc = formatResetTime(resetTime)
			fiveHourInfo.ResetTime = "resets " + resetDesc
		}
		
		// Display format: "95% left (resets 05:09)"
		if fiveHourInfo.ResetTime != "" {
			fiveHourInfo.Display = fmt.Sprintf("%d%% left (%s)", remaining, fiveHourInfo.ResetTime)
		} else {
			fiveHourInfo.Display = fmt.Sprintf("%d%% left", remaining)
		}
	}

	// Parse secondary (weekly limit) - store remaining percentage
	var weeklyInfo LimitInfo
	if resp.RateLimits.Secondary != nil {
		used := int(resp.RateLimits.Secondary.UsedPercent)
		remaining := 100 - used
		if remaining < 0 {
			remaining = 0
		}
		weeklyInfo.Percentage = remaining // Store remaining, not used
		
		resetDesc := ""
		if resp.RateLimits.Secondary.ResetsAt > 0 {
			resetTime := time.Unix(resp.RateLimits.Secondary.ResetsAt, 0)
			resetDesc = formatResetTimeWithDate(resetTime)
			weeklyInfo.ResetTime = "resets " + resetDesc
		}
		
		// Display format: "98% left (resets 16:22 on 10 Feb)"
		if weeklyInfo.ResetTime != "" {
			weeklyInfo.Display = fmt.Sprintf("%d%% left (%s)", remaining, weeklyInfo.ResetTime)
		} else {
			weeklyInfo.Display = fmt.Sprintf("%d%% left", remaining)
		}
	}

	// Use 5h limit as primary
	primaryPercent := fiveHourInfo.Percentage
	if resp.RateLimits.Primary == nil && resp.RateLimits.Secondary != nil {
		primaryPercent = weeklyInfo.Percentage
	}

	// Determine color based on remaining percentage (high remaining = green)
	color := "green"
	if primaryPercent <= 20 {
		color = "red"
	} else if primaryPercent <= 40 {
		color = "yellow"
	}

	return UsageInfo{
		Percentage:    primaryPercent,
		Display:       fiveHourInfo.Display,
		Color:         color,
		Source:        "rpc",
		LastFetched:   now,
		FiveHourLimit: fiveHourInfo,
		WeeklyLimit:   weeklyInfo,
	}, nil
}

// formatResetTime formats a reset time for 5h limit (time only).
func formatResetTime(t time.Time) string {
	return t.Format("15:04")
}

// formatResetTimeWithDate formats a reset time for weekly limit (time + date).
func formatResetTimeWithDate(t time.Time) string {
	return t.Format("15:04 2 Jan")
}
