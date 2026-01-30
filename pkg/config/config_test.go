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
	if len(tools) != 3 {
		t.Errorf("Expected 3 tools, got %d", len(tools))
	}
	
	// Check that all expected tools are present
	expectedTools := []string{"claude", "copilot", "codex"}
	for _, name := range expectedTools {
		tool := registry.Get(name)
		if tool == nil {
			t.Errorf("Tool %s not found in registry", name)
			continue
		}
		
		// Verify install commands exist
		if len(tool.InstallCmds) == 0 {
			t.Errorf("Tool %s has no install commands", name)
		}
		
		// Verify install URL exists
		if tool.InstallURL == "" {
			t.Errorf("Tool %s has no install URL", name)
		}
		
		// Check that install commands exist for all major platforms
		platforms := []string{"darwin", "linux", "windows"}
		for _, platform := range platforms {
			if tool.InstallCmds[platform] == "" {
				t.Errorf("Tool %s missing install command for %s", name, platform)
			}
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
