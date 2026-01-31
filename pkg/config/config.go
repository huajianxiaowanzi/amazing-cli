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
		DisplayName: "claude-code",
		Command:     "claude",
		Description: "Claude Code by Anthropic",
		Args:        []string{},
		InstallCmds: map[string]string{
			"darwin":  "curl -fsSL https://claude.ai/install.sh | bash",
			"linux":   "curl -fsSL https://claude.ai/install.sh | bash",
			"windows_ps":  "irm https://claude.ai/install.ps1 | iex",
			"windows_cmd": "curl -fsSL https://claude.ai/install.cmd -o install.cmd && install.cmd && del install.cmd",
		},
		InstallURL: "https://docs.anthropic.com/en/docs/claude-code/getting-started",
	})

	registry.Register(&tool.Tool{
		Name:        "copilot",
		DisplayName: "copilot",
		Command:     "copilot",
		Description: "GitHub's AI-powered CLI assistant",
		Args:        []string{},
		InstallCmds: map[string]string{
			"darwin": "(curl -fsSL https://gh.io/copilot-install | bash) || (wget -qO- https://gh.io/copilot-install | bash) || brew install copilot-cli || npm install -g @github/copilot || npm install -g @github/copilot@prerelease",
			"linux":  "(curl -fsSL https://gh.io/copilot-install | bash) || (wget -qO- https://gh.io/copilot-install | bash) || brew install copilot-cli || npm install -g @github/copilot || npm install -g @github/copilot@prerelease",
			"windows_ps": "winget install GitHub.Copilot; if ($LASTEXITCODE -ne 0) { npm install -g @github/copilot }; if ($LASTEXITCODE -ne 0) { npm install -g @github/copilot@prerelease }",
			"windows_cmd": "winget install GitHub.Copilot || npm install -g @github/copilot || npm install -g @github/copilot@prerelease",
		},
		InstallURL: "https://github.com/github/copilot-cli",
	})

	registry.Register(&tool.Tool{
		Name:        "kimi",
		DisplayName: "kimi",
		Command:     "kimi",
		Description: "Kimi Code by Moonshot",
		Args:        []string{},
		InstallCmds: map[string]string{
			"darwin":     "curl -L https://code.kimi.com/install.sh | bash",
			"linux":      "curl -L https://code.kimi.com/install.sh | bash",
			"windows_ps": "irm https://code.kimi.com/install.ps1 | iex",
		},
		InstallURL: "https://code.kimi.com",
	})

	registry.Register(&tool.Tool{
		Name:        "codex",
		DisplayName: "codex",
		Command:     "codex",
		Description: "OpenAI's Codex CLI",
		Args:        []string{},
		InstallCmds: map[string]string{
			"darwin":     "brew install codex || npm i -g @openai/codex",
			"linux":      "npm i -g @openai/codex",
			"windows_ps": "npm i -g @openai/codex",
			"windows_cmd": "npm i -g @openai/codex",
		},
		InstallURL: "https://platform.openai.com/docs/guides/code",
	})

	registry.Register(&tool.Tool{
		Name:        "opencode",
		DisplayName: "OpenCode ",
		Command:     "opencode",
		Description: "OpenCode ",
		Args:        []string{},
		InstallCmds: map[string]string{
			"darwin":      "brew install anomalyco/tap/opencode || curl -fsSL https://opencode.ai/install | bash",
			"linux":       "curl -fsSL https://opencode.ai/install | bash",
			"windows_ps":  "npm i -g opencode-ai",
			"windows_cmd": "npm i -g opencode-ai",
		},
		InstallURL: "https://opencode.ai",
	})

	return registry
}
