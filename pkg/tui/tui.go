// Package tui provides the terminal user interface using Bubble Tea.
package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/huajianxiaowanzi/amazing-cli/pkg/config"
	"github.com/huajianxiaowanzi/amazing-cli/pkg/tool"
)

// Styles for the TUI
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")).
			MarginTop(1).
			MarginBottom(1)

	selectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			PaddingLeft(1).
			PaddingRight(1)

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			PaddingLeft(1).
			PaddingRight(1)

	descStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			Italic(true)

	installedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575"))

	notInstalledStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF6B6B"))

	balanceStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")).
			Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			MarginTop(1)
)

// Model represents the TUI state.
type Model struct {
	tools    []*tool.Tool
	cursor   int
	selected string
	balance  config.Balance
	quitting bool
	err      error
}

// NewModel creates a new TUI model with the given tool registry.
func NewModel(registry *tool.Registry) Model {
	return Model{
		tools:   registry.List(),
		cursor:  0,
		balance: config.GetDefaultBalance(),
	}
}

// Init initializes the model (required by Bubble Tea).
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model (required by Bubble Tea).
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.tools)-1 {
				m.cursor++
			}

		case "enter":
			// User selected a tool
			m.selected = m.tools[m.cursor].Name
			return m, tea.Quit
		}
	}

	return m, nil
}

// View renders the TUI (required by Bubble Tea).
func (m Model) View() string {
	if m.quitting {
		return ""
	}

	var s strings.Builder

	// Title
	s.WriteString(titleStyle.Render("🚀 Amazing CLI - Select Your AI Tool"))
	s.WriteString("\n\n")

	// Tool list
	for i, t := range m.tools {
		cursor := "  "
		style := normalStyle
		if m.cursor == i {
			cursor = "▶ "
			style = selectedStyle
		}

		// Check if tool is installed
		statusIcon := notInstalledStyle.Render("✗")
		if t.IsInstalled() {
			statusIcon = installedStyle.Render("✓")
		}

		// Render tool item with inline token balance
		toolName := style.Render(t.DisplayName)
		balanceBar := renderInlineBalanceBar(m.balance)
		s.WriteString(fmt.Sprintf("%s%s %s  %s\n", cursor, statusIcon, toolName, balanceBar))
	}

	// Help text
	s.WriteString("\n")
	s.WriteString(helpStyle.Render("↑/↓: navigate • enter: launch • q: quit"))

	return s.String()
}

// GetSelected returns the name of the selected tool, if any.
func (m Model) GetSelected() string {
	return m.selected
}

// renderInlineBalanceBar creates a compact visual representation of the token balance.
func renderInlineBalanceBar(balance config.Balance) string {
	width := 15
	
	// Clamp percentage to 0-100 range
	percentage := balance.Percentage
	if percentage < 0 {
		percentage = 0
	}
	if percentage > 100 {
		percentage = 100
	}
	
	filled := (width * percentage) / 100
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	
	style := lipgloss.NewStyle()
	switch balance.Color {
	case "green":
		style = style.Foreground(lipgloss.Color("#04B575"))
	case "yellow":
		style = style.Foreground(lipgloss.Color("#FFB86C"))
	case "red":
		style = style.Foreground(lipgloss.Color("#FF6B6B"))
	default:
		// Default to green for unknown colors
		style = style.Foreground(lipgloss.Color("#04B575"))
	}

	label := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Render(fmt.Sprintf("Token: %s", balance.Display))
	
	return fmt.Sprintf("%s %s", label, style.Render(bar))
}

// Run starts the TUI and returns the selected tool name.
func Run(registry *tool.Registry) (string, error) {
	model := NewModel(registry)
	p := tea.NewProgram(model)
	
	finalModel, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("error running TUI: %w", err)
	}

	m, ok := finalModel.(Model)
	if !ok {
		return "", fmt.Errorf("unexpected model type returned from TUI")
	}
	return m.GetSelected(), nil
}
