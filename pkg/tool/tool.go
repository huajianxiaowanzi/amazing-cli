// Package tool provides interfaces and types for managing AI CLI tools.
package tool

import (
	"fmt"
	"os/exec"
	"syscall"
)

// Tool represents an AI CLI tool that can be launched.
type Tool struct {
	Name        string // Internal identifier (e.g., "aider")
	DisplayName string // Human-readable name (e.g., "Aider - AI Pair Programming")
	Command     string // Command to execute (e.g., "aider")
	Description string // Brief description of the tool
	Args        []string // Default arguments to pass
}

// IsInstalled checks if the tool is available on the system.
func (t *Tool) IsInstalled() bool {
	_, err := exec.LookPath(t.Command)
	return err == nil
}

// Execute launches the tool using syscall.Exec to replace the current process.
// This allows the tool to take over the terminal completely.
func (t *Tool) Execute() error {
	path, err := exec.LookPath(t.Command)
	if err != nil {
		return fmt.Errorf("tool not found: %s", t.Command)
	}

	// Prepare arguments (command name should be first)
	args := append([]string{t.Command}, t.Args...)

	// Replace current process with the tool
	// This is the cleanest way to hand off control to another CLI
	return syscall.Exec(path, args, nil)
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

// List returns all registered tools.
func (r *Registry) List() []*Tool {
	return r.tools
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
