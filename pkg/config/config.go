// Package config provides configuration for available AI tools.
package config

import (
	"github.com/huajianxiaowanzi/amazing-cli/pkg/tool"
)

// Balance represents a placeholder for token/credit balance information.
// This is designed to be extensible for future balance tracking implementations.
type Balance struct {
	Percentage int    // 0-100, current placeholder shows 100%
	Display    string // Human-readable display (e.g., "100%", "1000 tokens")
	Color      string // Color hint for display (e.g., "green", "yellow", "red")
}

// GetDefaultBalance returns the default placeholder balance.
// In the future, this can be replaced with actual API calls to check balances.
func GetDefaultBalance() Balance {
	return Balance{
		Percentage: 100,
		Display:    "100%",
		Color:      "green",
	}
}

// BalanceProvider defines the interface for balance checking.
// Implementations can query actual API endpoints for real balance data.
type BalanceProvider interface {
	GetBalance(toolName string) (Balance, error)
}

// LoadDefaultTools returns a registry with pre-configured AI tools.
func LoadDefaultTools() *tool.Registry {
	registry := tool.NewRegistry()

	// Register supported AI CLI tools
	// Note: Installation commands should be verified and updated based on actual installation methods
	registry.Register(&tool.Tool{
		Name:        "claude",
		DisplayName: "Claude Code",
		Command:     "claude",
		Description: "Claude Code by Anthropic",
		Args:        []string{},
		InstallCmds: map[string]string{
			// Note: These commands may need to be updated based on actual Claude CLI availability
			"darwin":  "", // Auto-install not available
			"linux":   "", // Auto-install not available
			"windows": "", // Auto-install not available
		},
		InstallURL: "https://docs.anthropic.com/claude/docs",
	})

	registry.Register(&tool.Tool{
		Name:        "copilot",
		DisplayName: "Copilot CLI",
		Command:     "github-copilot-cli",
		Description: "GitHub's AI-powered CLI assistant",
		Args:        []string{},
		InstallCmds: map[string]string{
			"darwin":  "gh extension install github/gh-copilot",
			"linux":   "gh extension install github/gh-copilot",
			"windows": "gh extension install github/gh-copilot",
		},
		InstallURL: "https://github.com/github/copilot-cli",
	})

	registry.Register(&tool.Tool{
		Name:        "aider",
		DisplayName: "Aider",
		Command:     "aider",
		Description: "AI pair programming in your terminal",
		Args:        []string{},
		InstallCmds: map[string]string{
			"darwin":  "pip install aider-chat",
			"linux":   "pip install aider-chat",
			"windows": "pip install aider-chat",
		},
		InstallURL: "https://aider.chat",
	})

	registry.Register(&tool.Tool{
		Name:        "codex",
		DisplayName: "Codex",
		Command:     "codex",
		Description: "OpenAI's Codex CLI",
		Args:        []string{},
		InstallCmds: map[string]string{
			// Note: These commands may need to be updated based on actual Codex CLI availability
			"darwin":  "", // Auto-install not available
			"linux":   "", // Auto-install not available
			"windows": "", // Auto-install not available
		},
		InstallURL: "https://platform.openai.com/docs/guides/code",
	})

	return registry
}
