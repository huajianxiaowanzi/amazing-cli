// Amazing CLI - A beautiful TUI launcher for AI agent command-line tools.
package main

import (
	"fmt"
	"os"

	"github.com/huajianxiaowanzi/amazing-cli/pkg/config"
	"github.com/huajianxiaowanzi/amazing-cli/pkg/tui"
)

func main() {
	// Load available AI tools
	registry := config.LoadDefaultTools()

	// Run the TUI and get user selection
	selectedTool, err := tui.Run(registry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// If user quit without selecting, exit gracefully
	if selectedTool == "" {
		os.Exit(0)
	}

	// Get the selected tool
	tool := registry.Get(selectedTool)
	if tool == nil {
		fmt.Fprintf(os.Stderr, "Error: tool not found: %s\n", selectedTool)
		os.Exit(1)
	}

	// Execute the tool (replaces current process)
	// This allows the tool to take full control of the terminal
	// Note: The TUI now handles installation prompts, so we only get here if tool is installed
	err = tool.Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing tool: %v\n", err)
		os.Exit(1)
	}
}
