package config

import (
	"testing"
)

func TestLoadDefaultTools(t *testing.T) {
	registry := LoadDefaultTools()

	if registry == nil {
		t.Fatal("LoadDefaultTools() returned nil")
	}

	tools := registry.List()
	if len(tools) != 5 {
		t.Errorf("Expected 5 tools, got %d", len(tools))
	}

	// Check that all expected tools are present
	expectedTools := []string{"claude", "copilot", "kimi", "codex", "opencode"}
	for _, name := range expectedTools {
		tool := registry.Get(name)
		if tool == nil {
			t.Errorf("Tool %s not found in registry", name)
			continue
		}

		// Verify install URL exists
		if tool.InstallURL == "" {
			t.Errorf("Tool %s has no install URL", name)
		}

		// Check that InstallCmds map exists (commands may be empty if auto-install not available)
		if tool.InstallCmds == nil {
			t.Errorf("Tool %s has nil InstallCmds map", name)
		}
	}
}

func TestGetDefaultBalance(t *testing.T) {
	balance := GetDefaultBalance()

	if balance.Percentage != 100 {
		t.Errorf("Expected percentage 100, got %d", balance.Percentage)
	}

	if balance.Display != "100%" {
		t.Errorf("Expected display '100%%', got %s", balance.Display)
	}

	if balance.Color != "green" {
		t.Errorf("Expected color 'green', got %s", balance.Color)
	}
}
