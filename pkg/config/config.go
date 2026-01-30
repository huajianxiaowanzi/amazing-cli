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

	// Register popular AI CLI tools
	registry.Register(&tool.Tool{
		Name:        "aider",
		DisplayName: "Aider - AI Pair Programming",
		Command:     "aider",
		Description: "AI pair programming in your terminal",
		Args:        []string{},
	})

	registry.Register(&tool.Tool{
		Name:        "claude",
		DisplayName: "Claude - Anthropic AI Assistant",
		Command:     "claude",
		Description: "Conversational AI by Anthropic",
		Args:        []string{},
	})

	registry.Register(&tool.Tool{
		Name:        "copilot",
		DisplayName: "GitHub Copilot CLI",
		Command:     "github-copilot-cli",
		Description: "GitHub's AI-powered CLI assistant",
		Args:        []string{},
	})

	registry.Register(&tool.Tool{
		Name:        "openai",
		DisplayName: "OpenAI CLI",
		Command:     "openai",
		Description: "OpenAI command-line interface",
		Args:        []string{},
	})

	return registry
}
