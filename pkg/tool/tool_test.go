package tool

import (
	"runtime"
	"testing"
)

func TestTool_HasInstallCommand(t *testing.T) {
	tests := []struct {
		name     string
		tool     *Tool
		expected bool
	}{
		{
			name: "Tool with install commands for current OS",
			tool: &Tool{
				Name:    "test-tool",
				Command: "test",
				InstallCmds: map[string]string{
					"darwin":  "brew install test",
					"linux":   "apt-get install test",
					"windows": "choco install test",
				},
			},
			expected: true,
		},
		{
			name: "Tool without install commands",
			tool: &Tool{
				Name:        "test-tool",
				Command:     "test",
				InstallCmds: map[string]string{},
			},
			expected: false,
		},
		{
			name: "Tool with empty install command for current OS",
			tool: &Tool{
				Name:    "test-tool",
				Command: "test",
				InstallCmds: map[string]string{
					runtime.GOOS: "",
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.tool.HasInstallCommand()
			if got != tt.expected {
				t.Errorf("HasInstallCommand() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestTool_Install_NoCommand(t *testing.T) {
	tool := &Tool{
		Name:        "test-tool",
		Command:     "test",
		InstallCmds: map[string]string{},
	}

	err := tool.Install()
	if err == nil {
		t.Error("Install() should return error when no install command available")
	}
}

func TestRegistry_Get(t *testing.T) {
	registry := NewRegistry()
	
	tool1 := &Tool{Name: "tool1", Command: "cmd1"}
	tool2 := &Tool{Name: "tool2", Command: "cmd2"}
	
	registry.Register(tool1)
	registry.Register(tool2)
	
	// Test getting existing tool
	got := registry.Get("tool1")
	if got == nil {
		t.Error("Get() should return tool1")
	}
	if got != nil && got.Name != "tool1" {
		t.Errorf("Get() returned wrong tool, got %v", got.Name)
	}
	
	// Test getting non-existent tool
	got = registry.Get("nonexistent")
	if got != nil {
		t.Error("Get() should return nil for non-existent tool")
	}
}

func TestRegistry_List(t *testing.T) {
	registry := NewRegistry()
	
	if len(registry.List()) != 0 {
		t.Error("New registry should have no tools")
	}
	
	tool1 := &Tool{Name: "tool1"}
	tool2 := &Tool{Name: "tool2"}
	
	registry.Register(tool1)
	registry.Register(tool2)
	
	tools := registry.List()
	if len(tools) != 2 {
		t.Errorf("Registry should have 2 tools, got %d", len(tools))
	}
}
