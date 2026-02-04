// Package codex provides functionality to fetch Codex token usage information.
package codex

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	// defaultWaitForOutputMs is the default time to wait for CLI output in milliseconds
	defaultWaitForOutputMs = 1500
)

// UsageInfo represents Codex token usage information.
type UsageInfo struct {
	Percentage   int       // 0-100, percentage used
	Display      string    // Human-readable display (e.g., "45%", "2h 30m remaining")
	Color        string    // Color hint: "green", "yellow", "red"
	ResetTime    time.Time // When the limit resets
	LastFetched  time.Time // When this data was fetched
	Source       string    // Where this data came from: "cli", "oauth", "cache"
	ErrorMessage string    // Error message if fetch failed
}

// OAuthCredentials represents the OAuth tokens stored in ~/.codex/auth.json
type OAuthCredentials struct {
	Tokens struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		IDToken      string `json:"id_token"`
		AccountID    string `json:"account_id"`
	} `json:"tokens"`
	LastRefresh time.Time `json:"last_refresh"`
}

// UsageFetcher provides methods to fetch Codex token usage.
type UsageFetcher struct {
	cacheFile string
	cacheTTL  time.Duration
}

// NewUsageFetcher creates a new UsageFetcher.
func NewUsageFetcher() *UsageFetcher {
	homeDir, _ := os.UserHomeDir()
	cacheDir := filepath.Join(homeDir, ".amazing-cli", "cache")
	os.MkdirAll(cacheDir, 0755)

	return &UsageFetcher{
		cacheFile: filepath.Join(cacheDir, "codex-usage.json"),
		cacheTTL:  5 * time.Minute, // Cache for 5 minutes
	}
}

// GetUsage fetches the current Codex token usage.
// It tries multiple strategies in order: OAuth, CLI PTY, Cache.
func (f *UsageFetcher) GetUsage(ctx context.Context) UsageInfo {
	// Try to load from cache first if it's fresh
	if cached, err := f.loadCache(); err == nil {
		if time.Since(cached.LastFetched) < f.cacheTTL {
			cached.Source = "cache"
			return cached
		}
	}

	// Try OAuth strategy (reading from ~/.codex/auth.json)
	if usage, err := f.fetchFromOAuth(ctx); err == nil {
		f.saveCache(usage)
		return usage
	}

	// Try CLI PTY strategy (running codex /status)
	if usage, err := f.fetchFromCLI(ctx); err == nil {
		f.saveCache(usage)
		return usage
	}

	// If all strategies fail, return a default "unknown" state
	return UsageInfo{
		Percentage:   100, // Show full as fallback
		Display:      "100%",
		Color:        "green",
		Source:       "default",
		LastFetched:  time.Now(),
		ErrorMessage: "unable to fetch usage data",
	}
}

// fetchFromOAuth attempts to read OAuth credentials and fetch usage.
// This is a simplified version - full implementation would need to handle token refresh
// and make API calls to ChatGPT backend.
func (f *UsageFetcher) fetchFromOAuth(ctx context.Context) (UsageInfo, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return UsageInfo{}, fmt.Errorf("failed to get home directory: %w", err)
	}

	authFile := filepath.Join(homeDir, ".codex", "auth.json")
	data, err := os.ReadFile(authFile)
	if err != nil {
		return UsageInfo{}, fmt.Errorf("failed to read auth file: %w", err)
	}

	var creds OAuthCredentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return UsageInfo{}, fmt.Errorf("failed to parse auth file: %w", err)
	}

	// TODO: Implement actual OAuth API calls
	// For now, return an error to fall back to CLI strategy
	return UsageInfo{}, fmt.Errorf("OAuth strategy not fully implemented")
}

// fetchFromCLI attempts to run "codex /status" and parse the output.
func (f *UsageFetcher) fetchFromCLI(ctx context.Context) (UsageInfo, error) {
	// Check if codex is installed
	codexPath, err := exec.LookPath("codex")
	if err != nil {
		return UsageInfo{}, fmt.Errorf("codex CLI not found: %w", err)
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	// Run codex with /status command
	// We need to send "/status\n" to the codex CLI
	cmd := exec.CommandContext(ctx, codexPath, "-s", "read-only", "-a", "untrusted")

	// Create pipes for stdin and stdout
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return UsageInfo{}, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stdout

	// Start the command
	if err := cmd.Start(); err != nil {
		return UsageInfo{}, fmt.Errorf("failed to start codex: %w", err)
	}

	// Send /status command
	if _, err := stdin.Write([]byte("/status\n")); err != nil {
		stdin.Close()
		cmd.Process.Kill()
		return UsageInfo{}, fmt.Errorf("failed to send /status command: %w", err)
	}
	stdin.Close()

	// Wait for output with a reasonable timeout
	// Use a smaller initial wait and check for completion
	outputChan := make(chan string, 1)
	go func() {
		time.Sleep(time.Duration(defaultWaitForOutputMs) * time.Millisecond)
		outputChan <- stdout.String()
	}()

	var output string
	select {
	case output = <-outputChan:
		// Got output, proceed
	case <-ctx.Done():
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		return UsageInfo{}, fmt.Errorf("timeout waiting for codex output")
	}

	// Kill the process (codex CLI stays running)
	if cmd.Process != nil {
		cmd.Process.Kill()
	}

	// Parse the output
	return parseStatusOutput(output)
}

// parseStatusOutput parses the output of "codex /status" command.
// It looks for patterns like:
// - "5h limit: 45% used (resets in 2h 30m)"
// - "Weekly limit: 23% used (resets in 4 days)"
// - "Credits: 1,234.56"
func parseStatusOutput(output string) (UsageInfo, error) {
	scanner := bufio.NewScanner(strings.NewReader(output))

	var fiveHourPercent int
	var fiveHourReset string
	var weeklyPercent int
	found := false

	// Regex patterns
	// Match patterns like "45% used" or "45.5% used"
	usedPattern := regexp.MustCompile(`(\d+(?:\.\d+)?)\s*%\s*used`)
	// Match patterns like "resets in 2h 30m" or "resets in 4 days"
	resetPattern := regexp.MustCompile(`resets in (.+)`)

	for scanner.Scan() {
		line := scanner.Text()

		// Look for 5h limit line
		if strings.Contains(line, "5h limit") || strings.Contains(line, "5-hour") {
			if matches := usedPattern.FindStringSubmatch(line); len(matches) > 1 {
				if percent, err := strconv.ParseFloat(matches[1], 64); err == nil {
					fiveHourPercent = int(percent)
					found = true
				}
			}
			if matches := resetPattern.FindStringSubmatch(line); len(matches) > 1 {
				fiveHourReset = matches[1]
			}
		}

		// Look for weekly limit line (as backup)
		if strings.Contains(line, "Weekly limit") || strings.Contains(line, "weekly") {
			if matches := usedPattern.FindStringSubmatch(line); len(matches) > 1 {
				if percent, err := strconv.ParseFloat(matches[1], 64); err == nil {
					weeklyPercent = int(percent)
					if !found { // Only use weekly if we didn't find 5h limit
						found = true
					}
				}
			}
		}
	}

	if !found {
		return UsageInfo{}, fmt.Errorf("failed to parse usage from codex output")
	}

	// Use 5h limit as primary, fall back to weekly
	usedPercent := fiveHourPercent
	if usedPercent == 0 {
		usedPercent = weeklyPercent
	}

	// Determine color based on usage
	color := "green"
	if usedPercent >= 80 {
		color = "red"
	} else if usedPercent >= 60 {
		color = "yellow"
	}

	// Build display string
	display := fmt.Sprintf("%d%%", usedPercent)
	if fiveHourReset != "" {
		display = fmt.Sprintf("%d%% (%s)", usedPercent, fiveHourReset)
	}

	return UsageInfo{
		Percentage:  usedPercent,
		Display:     display,
		Color:       color,
		Source:      "cli",
		LastFetched: time.Now(),
	}, nil
}

// ParseStatusOutputForTest is an exported version of parseStatusOutput for testing purposes.
func ParseStatusOutputForTest(output string) (UsageInfo, error) {
	return parseStatusOutput(output)
}

// loadCache loads cached usage info from disk.
func (f *UsageFetcher) loadCache() (UsageInfo, error) {
	data, err := os.ReadFile(f.cacheFile)
	if err != nil {
		return UsageInfo{}, err
	}

	var info UsageInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return UsageInfo{}, err
	}

	return info, nil
}

// saveCache saves usage info to disk cache.
func (f *UsageFetcher) saveCache(info UsageInfo) error {
	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(f.cacheFile, data, 0644)
}
