package codex

import (
	"fmt"
	"strings"
	"testing"
)

func TestParseStatusOutput(t *testing.T) {
	tests := []struct {
		name           string
		output         string
		expectError    bool
		expectPercent  int
		expectColor    string
		expectContains string
	}{
		{
			name: "5h limit with reset time",
			output: `
Welcome to Codex
5h limit: 45% used (resets in 2h 30m)
Weekly limit: 10% used (resets in 4 days)
Credits: 1,234.56
`,
			expectError:    false,
			expectPercent:  45,
			expectColor:    "green",
			expectContains: "2h 30m",
		},
		{
			name: "high usage - red color",
			output: `
5h limit: 85% used (resets in 1h)
Weekly limit: 20% used
`,
			expectError:   false,
			expectPercent: 85,
			expectColor:   "red",
		},
		{
			name: "medium usage - yellow color",
			output: `
5h limit: 65% used (resets in 3h)
`,
			expectError:   false,
			expectPercent: 65,
			expectColor:   "yellow",
		},
		{
			name: "weekly limit only",
			output: `
Weekly limit: 30% used (resets in 3 days)
`,
			expectError:   false,
			expectPercent: 30,
			expectColor:   "green",
		},
		{
			name: "no usage data",
			output: `
Welcome to Codex
Type /help for assistance
`,
			expectError: true,
		},
		{
			name: "decimal percentage",
			output: `
5h limit: 42.5% used (resets in 1h 15m)
`,
			expectError:   false,
			expectPercent: 42,
			expectColor:   "green",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseStatusOutput(tt.output)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result.Percentage != tt.expectPercent {
				t.Errorf("expected percentage %d, got %d", tt.expectPercent, result.Percentage)
			}

			if result.Color != tt.expectColor {
				t.Errorf("expected color %s, got %s", tt.expectColor, result.Color)
			}

			if tt.expectContains != "" && !strings.Contains(result.Display, tt.expectContains) {
				t.Errorf("expected display to contain %q, got %q", tt.expectContains, result.Display)
			}

			if result.Source != "cli" {
				t.Errorf("expected source to be 'cli', got %s", result.Source)
			}
		})
	}
}

func TestUsageInfoColorMapping(t *testing.T) {
	tests := []struct {
		percentage    int
		expectedColor string
	}{
		{0, "green"},
		{30, "green"},
		{59, "green"},
		{60, "yellow"},
		{75, "yellow"},
		{79, "yellow"},
		{80, "red"},
		{95, "red"},
		{100, "red"},
	}

	for _, tt := range tests {
		output := fmt.Sprintf("5h limit: %d%% used\n", tt.percentage)
		result, err := parseStatusOutput(output)

		if err != nil {
			t.Errorf("for %d%%, unexpected error: %v", tt.percentage, err)
			continue
		}

		if result.Color != tt.expectedColor {
			t.Errorf("for %d%%, expected color %s, got %s", tt.percentage, tt.expectedColor, result.Color)
		}
	}
}
