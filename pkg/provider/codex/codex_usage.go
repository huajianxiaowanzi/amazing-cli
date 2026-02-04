// Package codex provides functionality to fetch Codex token usage information.
package codex

import (
	"bufio"
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
	defaultWaitForOutputMs = 3000
	// maxWaitForOutputMs is the hard timeout for CLI output collection
	maxWaitForOutputMs = 10000
)

// LimitInfo represents information about a single limit (5h or weekly).
type LimitInfo struct {
	Percentage int    // 0-100, percentage used
	Display    string // Human-readable display (e.g., "0% (resets 03:31 5 Feb)")
	ResetTime  string // When the limit resets
}

// UsageInfo represents Codex token usage information.
type UsageInfo struct {
	Percentage   int       // 0-100, percentage used (from primary limit)
	Display      string    // Human-readable display (e.g., "45%", "2h 30m remaining")
	Color        string    // Color hint: "green", "yellow", "red"
	ResetTime    time.Time // When the limit resets
	LastFetched  time.Time // When this data was fetched
	Source       string    // Where this data came from: "cli", "oauth", "cache"
	ErrorMessage string    // Error message if fetch failed
	
	// Individual limit information
	FiveHourLimit LimitInfo // 5h limit details
	WeeklyLimit   LimitInfo // Weekly limit details
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
// It tries multiple strategies in order: OAuth API, RPC, CLI PTY.
// Priority: OAuth API (fastest) > RPC > CLI PTY
func (f *UsageFetcher) GetUsage(ctx context.Context) UsageInfo {
	// Try to load from cache first if it's fresh
	if cached, err := f.loadCache(); err == nil {
		if time.Since(cached.LastFetched) < f.cacheTTL {
			cached.Source = "cache"
			return cached
		}
	}

	// Try OAuth API strategy (fastest, most accurate) - Priority 1
	if usage, err := FetchUsageViaOAuth(ctx); err == nil {
		f.saveCache(usage)
		return usage
	}

	// Try RPC strategy (codex app-server) - Priority 2
	if usage, err := FetchUsageViaRPC(ctx); err == nil {
		f.saveCache(usage)
		return usage
	}

	// Try CLI PTY strategy (running codex /status) as fallback - Priority 3
	if usage, err := f.fetchFromCLI(ctx); err == nil {
		f.saveCache(usage)
		return usage
	}

	// If all strategies fail, return a default "unknown" state with dual limits
	return UsageInfo{
		Percentage:   0, // Show 0% as fallback (unknown)
		Display:      "?%",
		Color:        "green",
		Source:       "default",
		LastFetched:  time.Now(),
		ErrorMessage: "unable to fetch usage data",
		FiveHourLimit: LimitInfo{
			Percentage: 0,
			Display:    "?%",
			ResetTime:  "",
		},
		WeeklyLimit: LimitInfo{
			Percentage: 0,
			Display:    "?%",
			ResetTime:  "",
		},
	}
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

	output, err := runCodexStatus(ctx, codexPath)
	if err != nil {
		f.writeDebugOutput("runCodexStatus error", err.Error())
		return UsageInfo{}, err
	}

	// Parse the output
	usage, parseErr := parseStatusOutput(output)
	if parseErr != nil {
		f.writeDebugOutput("parseStatusOutput error", output)
		return UsageInfo{}, parseErr
	}
	return usage, nil
}

// parseStatusOutput parses the output of "codex /status" command.
// It looks for patterns like:
// Old format: "5h limit: 45% used (resets in 2h 30m)"
// New format: "5h limit: [████████████████████] 100% left (resets 03:31 on 5 Feb)"
// - "Weekly limit: 23% used (resets in 4 days)"
// - "Credits: 1,234.56"
func parseStatusOutput(output string) (UsageInfo, error) {
	cleanOutput := stripANSICodes(output)
	scanner := bufio.NewScanner(strings.NewReader(cleanOutput))

	var fiveHourPercent int
	var fiveHourReset string
	var weeklyPercent int
	var weeklyReset string
	foundFiveHour := false
	foundWeekly := false

	// Regex patterns
	// Match patterns like "45% used" or "45.5% used"
	usedPattern := regexp.MustCompile(`(\d+(?:\.\d+)?)\s*%\s*used`)
	// Match patterns like "100% left", "50% left", or "90% remaining"
	leftPattern := regexp.MustCompile(`(\d+(?:\.\d+)?)\s*%\s*(left|remaining)`)
	// Match patterns like "resets in 2h 30m" or "resets in 4 days"
	resetInPattern := regexp.MustCompile(`resets in (.+)`)
	// Match patterns like "resets 03:31 on 5 Feb" or "resets 16:22 on 10 Feb"
	resetOnPattern := regexp.MustCompile(`resets (\d{2}:\d{2}) on (\d+\s+\w+)`)
	// Match patterns like "resets 05:09"
	resetAtPattern := regexp.MustCompile(`resets (\d{2}:\d{2})`)

	for scanner.Scan() {
		line := scanner.Text()

		// Look for 5h limit line
		if strings.Contains(line, "5h limit") || strings.Contains(line, "5-hour") {
			// Try "% used" pattern first
			if matches := usedPattern.FindStringSubmatch(line); len(matches) > 1 {
				if percent, err := strconv.ParseFloat(matches[1], 64); err == nil {
					fiveHourPercent = int(percent)
					foundFiveHour = true
				}
			} else if matches := leftPattern.FindStringSubmatch(line); len(matches) > 1 {
				// Try "% left" pattern - convert to used
				if percent, err := strconv.ParseFloat(matches[1], 64); err == nil {
					fiveHourPercent = 100 - int(percent)
					foundFiveHour = true
				}
			}
			
			// Try to extract reset time
			if matches := resetInPattern.FindStringSubmatch(line); len(matches) > 1 {
				fiveHourReset = matches[1]
			} else if matches := resetOnPattern.FindStringSubmatch(line); len(matches) > 2 {
				fiveHourReset = fmt.Sprintf("%s %s", matches[1], matches[2])
			} else if matches := resetAtPattern.FindStringSubmatch(line); len(matches) > 1 {
				fiveHourReset = matches[1]
			}
		}

		// Look for weekly limit line
		if strings.Contains(line, "Weekly limit") || strings.Contains(line, "weekly") {
			// Try "% used" pattern first
			if matches := usedPattern.FindStringSubmatch(line); len(matches) > 1 {
				if percent, err := strconv.ParseFloat(matches[1], 64); err == nil {
					weeklyPercent = int(percent)
					foundWeekly = true
				}
			} else if matches := leftPattern.FindStringSubmatch(line); len(matches) > 1 {
				// Try "% left" pattern - convert to used
				if percent, err := strconv.ParseFloat(matches[1], 64); err == nil {
					weeklyPercent = 100 - int(percent)
					foundWeekly = true
				}
			}
			
			// Try to extract reset time
			if matches := resetInPattern.FindStringSubmatch(line); len(matches) > 1 {
				weeklyReset = matches[1]
			} else if matches := resetOnPattern.FindStringSubmatch(line); len(matches) > 2 {
				weeklyReset = fmt.Sprintf("%s %s", matches[1], matches[2])
			} else if matches := resetAtPattern.FindStringSubmatch(line); len(matches) > 1 {
				weeklyReset = matches[1]
			}
		}
	}

	if !foundFiveHour && !foundWeekly {
		return UsageInfo{}, fmt.Errorf("failed to parse usage from codex output")
	}

	// Use 5h limit as primary, fall back to weekly
	primaryPercent := fiveHourPercent
	primaryReset := fiveHourReset
	if !foundFiveHour && foundWeekly {
		primaryPercent = weeklyPercent
		primaryReset = weeklyReset
	}

	// Determine color based on primary usage
	color := "green"
	if primaryPercent >= 80 {
		color = "red"
	} else if primaryPercent >= 60 {
		color = "yellow"
	}

	// Build display string for primary limit
	display := fmt.Sprintf("%d%%", primaryPercent)
	if primaryReset != "" {
		display = fmt.Sprintf("%d%% (%s)", primaryPercent, primaryReset)
	}

	// Build LimitInfo structs
	fiveHourInfo := LimitInfo{
		Percentage: fiveHourPercent,
		ResetTime:  fiveHourReset,
	}
	if fiveHourReset != "" {
		fiveHourInfo.Display = fmt.Sprintf("%d%% (%s)", fiveHourPercent, fiveHourReset)
	} else {
		fiveHourInfo.Display = fmt.Sprintf("%d%%", fiveHourPercent)
	}

	weeklyInfo := LimitInfo{
		Percentage: weeklyPercent,
		ResetTime:  weeklyReset,
	}
	if weeklyReset != "" {
		weeklyInfo.Display = fmt.Sprintf("%d%% (%s)", weeklyPercent, weeklyReset)
	} else {
		weeklyInfo.Display = fmt.Sprintf("%d%%", weeklyPercent)
	}

	return UsageInfo{
		Percentage:    primaryPercent,
		Display:       display,
		Color:         color,
		Source:        "cli",
		LastFetched:   time.Now(),
		FiveHourLimit: fiveHourInfo,
		WeeklyLimit:   weeklyInfo,
	}, nil
}

func stripANSICodes(s string) string {
	// Strip common ANSI CSI escape sequences to make parsing robust.
	re := regexp.MustCompile(`\x1b\[[0-9;?]*[ -/]*[@-~]`)
	return re.ReplaceAllString(s, "")
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

func (f *UsageFetcher) writeDebugOutput(prefix, content string) {
	dir := filepath.Dir(f.cacheFile)
	_ = os.MkdirAll(dir, 0755)
	path := filepath.Join(dir, "codex-usage-debug.txt")
	_ = os.WriteFile(path, []byte(prefix+"\n"+content+"\n"), 0644)
}
