// Package codex provides functionality to fetch Codex token usage information.
package codex

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const (
	// chatGPTUsageURL is the endpoint for fetching Codex usage via OAuth
	chatGPTUsageURL = "https://chatgpt.com/backend-api/wham/usage"
)

// OAuthUsageResponse represents the response from the ChatGPT usage API.
type OAuthUsageResponse struct {
	PlanType  string           `json:"plan_type,omitempty"`
	RateLimit *RateLimitDetail `json:"rate_limit,omitempty"`
	Credits   *CreditDetail    `json:"credits,omitempty"`
}

// RateLimitDetail contains rate limit information.
type RateLimitDetail struct {
	PrimaryWindow   *WindowSnapshot `json:"primary_window,omitempty"`
	SecondaryWindow *WindowSnapshot `json:"secondary_window,omitempty"`
}

// WindowSnapshot represents a rate limit window.
type WindowSnapshot struct {
	UsedPercent        int `json:"used_percent"`
	ResetAt            int64 `json:"reset_at"`
	LimitWindowSeconds int `json:"limit_window_seconds"`
}

// CreditDetail contains credit information.
type CreditDetail struct {
	HasCredits bool    `json:"has_credits"`
	Unlimited  bool        `json:"unlimited"`
	Balance    json.Number `json:"balance,omitempty"` // Can be string or number in API response
}

// OAuthAuthFile represents the structure of ~/.codex/auth.json
type OAuthAuthFile struct {
	Tokens struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		IDToken      string `json:"id_token"`
		AccountID    string `json:"account_id"`
	} `json:"tokens"`
	LastRefresh string `json:"last_refresh"`
	// For API key mode
	OpenAIAPIKey string `json:"OPENAI_API_KEY,omitempty"`
}

// loadOAuthCredentials loads OAuth credentials from ~/.codex/auth.json
func loadOAuthCredentials() (*OAuthAuthFile, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	// Check CODEX_HOME environment variable first
	codexHome := os.Getenv("CODEX_HOME")
	if codexHome == "" {
		codexHome = filepath.Join(homeDir, ".codex")
	}

	authFile := filepath.Join(codexHome, "auth.json")
	data, err := os.ReadFile(authFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read auth file: %w", err)
	}

	var auth OAuthAuthFile
	if err := json.Unmarshal(data, &auth); err != nil {
		return nil, fmt.Errorf("failed to parse auth file: %w", err)
	}

	// Check if we have valid credentials
	if auth.Tokens.AccessToken == "" && auth.OpenAIAPIKey == "" {
		return nil, fmt.Errorf("no valid credentials found in auth.json")
	}

	return &auth, nil
}

// FetchUsageViaOAuth fetches usage information using OAuth API.
func FetchUsageViaOAuth(ctx context.Context) (UsageInfo, error) {
	creds, err := loadOAuthCredentials()
	if err != nil {
		return UsageInfo{}, err
	}

	// If using API key, OAuth API won't work
	if creds.OpenAIAPIKey != "" && creds.Tokens.AccessToken == "" {
		return UsageInfo{}, fmt.Errorf("API key mode does not support OAuth usage API")
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", chatGPTUsageURL, nil)
	if err != nil {
		return UsageInfo{}, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+creds.Tokens.AccessToken)
	req.Header.Set("User-Agent", "amazing-cli")
	req.Header.Set("Accept", "application/json")

	// Set account ID if available
	if creds.Tokens.AccountID != "" {
		req.Header.Set("ChatGPT-Account-Id", creds.Tokens.AccountID)
	}

	// Make request with timeout
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return UsageInfo{}, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return UsageInfo{}, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	switch resp.StatusCode {
	case http.StatusOK:
		// Success, parse response
	case http.StatusUnauthorized, http.StatusForbidden:
		return UsageInfo{}, fmt.Errorf("unauthorized: token may be expired, run 'codex' to re-authenticate")
	default:
		return UsageInfo{}, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var usageResp OAuthUsageResponse
	if err := json.Unmarshal(body, &usageResp); err != nil {
		return UsageInfo{}, fmt.Errorf("failed to parse response: %w", err)
	}

	return convertOAuthToUsageInfo(&usageResp)
}

// convertOAuthToUsageInfo converts OAuth API response to UsageInfo.
func convertOAuthToUsageInfo(resp *OAuthUsageResponse) (UsageInfo, error) {
	if resp.RateLimit == nil {
		return UsageInfo{}, fmt.Errorf("no rate limit data in response")
	}

	now := time.Now()

	// Parse primary window (5h limit) - store remaining percentage
	var fiveHourInfo LimitInfo
	if resp.RateLimit.PrimaryWindow != nil {
		used := resp.RateLimit.PrimaryWindow.UsedPercent
		remaining := 100 - used
		if remaining < 0 {
			remaining = 0
		}
		fiveHourInfo.Percentage = remaining // Store remaining, not used

		resetDesc := ""
		if resp.RateLimit.PrimaryWindow.ResetAt > 0 {
			resetTime := time.Unix(resp.RateLimit.PrimaryWindow.ResetAt, 0)
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

	// Parse secondary window (weekly limit) - store remaining percentage
	var weeklyInfo LimitInfo
	if resp.RateLimit.SecondaryWindow != nil {
		used := resp.RateLimit.SecondaryWindow.UsedPercent
		remaining := 100 - used
		if remaining < 0 {
			remaining = 0
		}
		weeklyInfo.Percentage = remaining // Store remaining, not used

		resetDesc := ""
		if resp.RateLimit.SecondaryWindow.ResetAt > 0 {
			resetTime := time.Unix(resp.RateLimit.SecondaryWindow.ResetAt, 0)
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
	if resp.RateLimit.PrimaryWindow == nil && resp.RateLimit.SecondaryWindow != nil {
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
		Source:        "oauth",
		LastFetched:   now,
		FiveHourLimit: fiveHourInfo,
		WeeklyLimit:   weeklyInfo,
	}, nil
}
