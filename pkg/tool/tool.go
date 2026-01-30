// Package tool provides interfaces and types for managing AI CLI tools.
package tool

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
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

	// Windows can provide separate PowerShell and CMD commands.
	if osType == "windows" {
		installCmdPS := t.InstallCmds["windows_ps"]
		installCmdCMD := t.InstallCmds["windows_cmd"]

		if installCmdPS != "" || installCmdCMD != "" {
			if installCmdPS != "" {
				if err := runInstallCommand(osType, installCmdPS, true); err == nil {
					return t.verifyInstalled()
				} else if installCmdCMD != "" {
					if err := runInstallCommand(osType, installCmdCMD, false); err != nil {
						return err
					}
					return t.verifyInstalled()
				} else {
					return err
				}
			}
			if err := runInstallCommand(osType, installCmdCMD, false); err != nil {
				return err
			}
			return t.verifyInstalled()
		}
	}

	// Check if we have installation commands for this OS
	installCmd, exists := t.InstallCmds[osType]
	if !exists || installCmd == "" {
		if t.InstallURL != "" {
			return fmt.Errorf("automated installation not available for %s. Please visit: %s", osType, t.InstallURL)
		}
		return fmt.Errorf("automated installation not available for %s", osType)
	}

	if err := runInstallCommand(osType, installCmd, true); err != nil {
		return err
	}
	return t.verifyInstalled()
}

// HasInstallCommand checks if the tool has an installation command for the current OS.
func (t *Tool) HasInstallCommand() bool {
	osType := runtime.GOOS
	if osType == "windows" {
		if t.InstallCmds["windows_ps"] != "" || t.InstallCmds["windows_cmd"] != "" {
			return true
		}
	}
	cmd, exists := t.InstallCmds[osType]
	return exists && cmd != ""
}

func runInstallCommand(osType, installCmd string, preferPowerShell bool) error {
	// Execute the installation command
	// Note: stdin is not connected to avoid race conditions with TUI
	var cmd *exec.Cmd
	if osType == "windows" {
		if preferPowerShell {
			cmd = exec.Command("powershell", "-Command", installCmd)
		} else {
			cmd = exec.Command("cmd", "/c", installCmd)
		}
	} else {
		cmd = exec.Command("sh", "-c", installCmd)
	}

	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output
	// stdin is intentionally not connected to prevent race conditions with TUI

	if err := cmd.Run(); err != nil {
		lastLine := lastNonEmptyLine(output.String())
		if lastLine != "" {
			return fmt.Errorf("install failed: %s", lastLine)
		}
		return fmt.Errorf("install failed")
	}
	return nil
}

func lastNonEmptyLine(s string) string {
	lines := strings.Split(s, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line != "" {
			return line
		}
	}
	return ""
}

func (t *Tool) verifyInstalled() error {
	if t.IsInstalled() {
		return nil
	}
	if runtime.GOOS != "windows" {
		if err := ensureLocalBinInPath(t.Command); err == nil {
			return nil
		}
	}
	return fmt.Errorf("install finished but %s is still not in PATH", t.Command)
}

func ensureLocalBinInPath(command string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	localBin := filepath.Join(home, ".local", "bin")
	target := filepath.Join(localBin, command)
	if _, err := os.Stat(target); err != nil {
		return err
	}

	if !pathContains(localBin) {
		if err := appendPathToShellConfig(localBin); err != nil {
			return err
		}
		_ = os.Setenv("PATH", localBin+string(os.PathListSeparator)+os.Getenv("PATH"))
	}

	_, err = exec.LookPath(command)
	return err
}

func pathContains(dir string) bool {
	for _, p := range filepath.SplitList(os.Getenv("PATH")) {
		if p == dir {
			return true
		}
	}
	return false
}

func appendPathToShellConfig(dir string) error {
	shell := filepath.Base(os.Getenv("SHELL"))
	var rc string
	switch shell {
	case "zsh":
		rc = ".zshrc"
	case "bash":
		rc = ".bashrc"
	default:
		return fmt.Errorf("unsupported shell: %s", shell)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	rcPath := filepath.Join(home, rc)
	line := fmt.Sprintf("export PATH=\"%s:$PATH\"\n", dir)

	if data, err := os.ReadFile(rcPath); err == nil {
		if strings.Contains(string(data), dir) {
			return nil
		}
	} else if !os.IsNotExist(err) {
		return err
	}

	f, err := os.OpenFile(rcPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(line)
	return err
}
