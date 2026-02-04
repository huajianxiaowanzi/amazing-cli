// Amazing CLI - A beautiful TUI launcher for AI agent command-line tools.
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/huajianxiaowanzi/amazing-cli/pkg/config"
	"github.com/huajianxiaowanzi/amazing-cli/pkg/provider/codex"
	"github.com/huajianxiaowanzi/amazing-cli/pkg/tool"
	"github.com/huajianxiaowanzi/amazing-cli/pkg/tui"
)

func main() {
	// Load available AI tools
	registry := config.LoadDefaultTools()

	// Load tool usage history
	usageData := config.LoadToolUsage()

	// Apply usage history to tools
	for _, t := range registry.List() {
		if lastUsed, ok := usageData[t.Name]; ok {
			t.LastUsed = lastUsed
		}
	}

	// Fetch balances for tools that support it
	fetchToolBalances(registry)

	// Run the TUI and get user selection
	selectedToolName, err := tui.Run(registry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// If user quit without selecting, exit gracefully
	if selectedToolName == "" {
		os.Exit(0)
	}

	// Get the selected tool
	selectedTool := registry.Get(selectedToolName)
	if selectedTool == nil {
		fmt.Fprintf(os.Stderr, "Error: tool not found: %s\n", selectedToolName)
		os.Exit(1)
	}

	// Safety check: verify tool is installed before execution
	// The TUI handles installation prompts, but we verify here as a safety measure
	if !selectedTool.IsInstalled() {
		fmt.Fprintf(os.Stderr, "\n❌ Tool not installed: %s\n", selectedTool.Command)
		fmt.Fprintf(os.Stderr, "Note: This should not happen if you used the TUI installation feature.\n")
		fmt.Fprintf(os.Stderr, "Please restart the application and try installing again.\n\n")
		os.Exit(1)
	}

	// Update usage data with current time
	usageData[selectedToolName] = time.Now()
	if err := config.SaveToolUsage(usageData); err != nil {
		// Non-fatal error, just log it
		fmt.Fprintf(os.Stderr, "Warning: failed to save usage data: %v\n", err)
	}

	// Execute the tool (replaces current process)
	// This allows the tool to take full control of the terminal
	err = selectedTool.Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing tool: %v\n", err)
		os.Exit(1)
	}
}

// fetchToolBalances fetches the balance for each tool that supports it.
func fetchToolBalances(registry *tool.Registry) {
	ctx := context.Background()

	for _, t := range registry.List() {
		// Only fetch for tools that are installed
		if !t.IsInstalled() {
			continue
		}

		// Fetch balance based on tool name
		switch t.Name {
		case "codex":
			fetcher := codex.NewBalanceFetcher()
			balance := fetcher.GetBalance(ctx)
			t.Balance = &tool.Balance{
				Percentage: balance.Percentage,
				Display:    balance.Display,
				Color:      balance.Color,
			}
		// Add more tools here as needed
		default:
			// Tools without specific balance fetchers get default balance
			continue
		}
	}
}
