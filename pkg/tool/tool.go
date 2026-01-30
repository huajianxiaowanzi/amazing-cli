// Package tool provides interfaces and types for managing AI CLI tools.
package tool

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

// Tool represents an AI CLI tool that can be launched.
type Tool struct {
	Name        string            // Internal identifier (e.g., "aider")
	DisplayName string            // Human-readable name (e.g., "Aider - AI Pair Programming")
	Command     string            // Command to execute (e.g., "aider")
	Description string            // Brief description of the tool
	Args        []string          // Default arguments to pass
	InstallCmds map[string]string // OS-specific installation commands (key: "windows", "darwin", "linux")
	InstallURL  string            // URL to installation documentation
}

// IsInstalled checks if the tool is available on the system.
func (t *Tool) IsInstalled() bool {
	_, err := exec.LookPath(t.Command)
	return err == nil
}

// clearScreen clears the terminal screen in a cross-platform way.
func clearScreen() {
	if runtime.GOOS == "windows" {
		// On Windows, use the cls command
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		// Ignore errors as clearing the screen is optional and shouldn't prevent tool execution
		_ = cmd.Run()
	} else {
		// On Unix-like systems, use ANSI escape sequences which are more reliable
		// \033[H moves cursor to home position, \033[2J clears the entire screen
		fmt.Print("\033[H\033[2J")
		// Flush to ensure the escape sequences are written immediately
		// Ignore errors as clearing the screen is optional and shouldn't prevent tool execution
		_ = os.Stdout.Sync()
	}
}

// Execute launches the tool as a child process with full terminal control.
// This method is cross-platform compatible (works on Windows, Linux, macOS).
func (t *Tool) Execute() error {
	path, err := exec.LookPath(t.Command)
	if err != nil {
		return fmt.Errorf("tool not found: %s", t.Command)
	}

	// Clear the screen before launching the tool
	clearScreen()

	// Create command with arguments
	cmd := exec.Command(path, t.Args...)

	// Pass through standard streams to allow full terminal interaction
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the command and wait for it to complete
	return cmd.Run()
}

// Registry manages a collection of available tools.
type Registry struct {
	tools []*Tool
}

// NewRegistry creates a new tool registry.
func NewRegistry() *Registry {
	return &Registry{
		tools: make([]*Tool, 0),
	}
}

// Register adds a tool to the registry.
func (r *Registry) Register(tool *Tool) {
	r.tools = append(r.tools, tool)
}

// List returns all registered tools sorted by installation status.
// Installed tools appear first, followed by uninstalled tools.
func (r *Registry) List() []*Tool {
	// Sort: installed tools first, then uninstalled
	// This preserves the registration order within each group
	var installed, uninstalled []*Tool
	for _, tool := range r.tools {
		if tool.IsInstalled() {
			installed = append(installed, tool)
		} else {
			uninstalled = append(uninstalled, tool)
		}
	}
	
	// Combine: installed first, then uninstalled
	result := make([]*Tool, 0, len(r.tools))
	result = append(result, installed...)
	result = append(result, uninstalled...)
	return result
}

// Get retrieves a tool by name.
func (r *Registry) Get(name string) *Tool {
	for _, tool := range r.tools {
		if tool.Name == name {
			return tool
		}
	}
	return nil
}

// Install attempts to install the tool on the current system.
// Returns an error if installation is not available or fails.
// Note: This method should not be called while a TUI is active, as it does not connect stdin
// to avoid race conditions between the TUI and installation process.
func (t *Tool) Install() error {
	osType := runtime.GOOS

	// Check if we have installation commands for this OS
	installCmd, exists := t.InstallCmds[osType]
	if !exists || installCmd == "" {
		if t.InstallURL != "" {
			return fmt.Errorf("automated installation not available for %s. Please visit: %s", osType, t.InstallURL)
		}
		return fmt.Errorf("automated installation not available for %s", osType)
	}

	// Execute the installation command
	// Note: stdin is not connected to avoid race conditions with TUI
	var cmd *exec.Cmd
	if osType == "windows" {
		cmd = exec.Command("powershell", "-Command", installCmd)
	} else {
		cmd = exec.Command("sh", "-c", installCmd)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// stdin is intentionally not connected to prevent race conditions with TUI

	return cmd.Run()
}

// HasInstallCommand checks if the tool has an installation command for the current OS.
func (t *Tool) HasInstallCommand() bool {
	osType := runtime.GOOS
	cmd, exists := t.InstallCmds[osType]
	return exists && cmd != ""
}
