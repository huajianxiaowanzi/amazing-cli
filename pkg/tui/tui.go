// Package tui provides the terminal user interface using Bubble Tea.
package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/huajianxiaowanzi/amazing-cli/pkg/config"
	"github.com/huajianxiaowanzi/amazing-cli/pkg/tool"
)

// installCompleteMsg is sent when installation completes
type installCompleteMsg struct {
	success bool
	err     error
}

// performInstall runs the installation in a goroutine
func performInstall(t *tool.Tool) tea.Cmd {
	return func() tea.Msg {
		err := t.Install()
		return installCompleteMsg{
			success: err == nil,
			err:     err,
		}
	}
}

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

	submenuStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#CFCFCF"))

	submenuSelectedStyle = lipgloss.NewStyle().
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

	dialogStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(1, 2).
			MarginTop(1).
			MarginBottom(1)

	warningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFB86C")).
			Bold(true)

	successMsgStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")).
			Bold(true)

	errorMsgStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6B6B")).
			Bold(true)
)

// Model represents the TUI state.
type Model struct {
	tools             []*tool.Tool
	cursor            int
	promptCursor      int
	spinner           spinner.Model
	selected          string
	balance           config.Balance
	quitting          bool
	err               error
	showInstallPrompt bool
	installing        bool
	installError      string
	installSuccess    bool
}

// NewModel creates a new TUI model with the given tool registry.
func NewModel(registry *tool.Registry) Model {
	spin := spinner.New()
	spin.Spinner = spinner.Line
	spin.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))
	return Model{
		tools:        registry.List(),
		cursor:       0,
		promptCursor: 0,
		spinner:      spin,
		balance:      config.GetDefaultBalance(),
	}
}

// Init initializes the model (required by Bubble Tea).
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model (required by Bubble Tea).
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case installCompleteMsg:
		m.installing = false
		if msg.success {
			m.installSuccess = true
			m.installError = ""
			// Refresh the tool's installation status by checking again
			// This updates the checkmark in the UI
		} else {
			m.installError = fmt.Sprintf("%v", msg.err)
		}
		return m, nil

	case tea.KeyMsg:
		// If showing install prompt
		if m.showInstallPrompt {
			switch msg.String() {
			case "up", "k":
				if m.promptCursor > 0 {
					m.promptCursor--
				}
				return m, nil
			case "down", "j":
				if m.promptCursor < 1 {
					m.promptCursor++
				}
				return m, nil
			case "enter", "y":
				selectedTool := m.tools[m.cursor]
				if m.promptCursor == 0 {
					// Start installation
					if selectedTool.HasInstallCommand() {
						m.installing = true
						m.showInstallPrompt = false
						return m, tea.Batch(performInstall(selectedTool), m.spinner.Tick)
					}
					if selectedTool.InstallURL != "" {
						m.installError = fmt.Sprintf("automated installation not available. Please visit: %s", selectedTool.InstallURL)
					} else {
						m.installError = "automated installation not available"
					}
					m.showInstallPrompt = false
					return m, nil
				}
				// Back
				m.showInstallPrompt = false
				m.installError = ""
				m.installSuccess = false
				return m, nil

			case "n", "q", "esc":
				// Cancel installation
				m.showInstallPrompt = false
				m.installError = ""
				m.installSuccess = false
				return m, nil
			}
			return m, nil
		}

		// If installation completed successfully, allow closing dialog
		if m.installSuccess {
			switch msg.String() {
			case "enter", "q", "esc":
				m.installSuccess = false
				return m, nil
			}
			return m, nil
		}

		// If there's an install error, allow closing dialog
		if m.installError != "" {
			switch msg.String() {
			case "enter", "q", "esc":
				m.installError = ""
				return m, nil
			}
			return m, nil
		}

		// Normal navigation
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
			selectedTool := m.tools[m.cursor]

			// Check if tool is installed
			if !selectedTool.IsInstalled() {
				// Show install prompt
				m.showInstallPrompt = true
				m.promptCursor = 0
				return m, nil
			}

			// Tool is installed, proceed to launch
			m.selected = selectedTool.Name
			return m, tea.Quit
		}
	}

	if m.installing {
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
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

		// Inline install options when tool is not installed and selected
		if m.showInstallPrompt && m.cursor == i && !t.IsInstalled() {
			installLabel := "自动安装"
			if !t.HasInstallCommand() {
				installLabel = "自动安装 (不可用)"
			}
			backLabel := "返回"

			installStyle := submenuStyle
			backStyle := submenuStyle
			if m.promptCursor == 0 {
				installStyle = submenuSelectedStyle
			} else {
				backStyle = submenuSelectedStyle
			}

			s.WriteString(fmt.Sprintf("   ├─ %s\n", installStyle.Render(installLabel)))
			s.WriteString(fmt.Sprintf("   └─ %s\n", backStyle.Render(backLabel)))
		}
	}

	// Show installation in progress
	if m.installing {
		s.WriteString("\n")
		var dialogContent strings.Builder
		dialogContent.WriteString(fmt.Sprintf("%s Installing...\n", m.spinner.View()))
		s.WriteString(dialogStyle.Render(dialogContent.String()))
		return s.String()
	}

	// Show installation success message
	if m.installSuccess {
		s.WriteString("\n")
		s.WriteString(successMsgStyle.Render("✓ Installed"))
		s.WriteString("\n")
		s.WriteString(helpStyle.Render("Press any key to continue"))
		return s.String()
	}

	// Show installation error message
	if m.installError != "" {
		s.WriteString("\n")
		s.WriteString(errorMsgStyle.Render("✗ Installation failed"))
		s.WriteString("\n")
		s.WriteString(descStyle.Render(m.installError))
		s.WriteString("\n")
		s.WriteString(helpStyle.Render("Press any key to continue"))
		return s.String()
	}

	// Help text
	s.WriteString("\n")
	if m.showInstallPrompt {
		s.WriteString(helpStyle.Render("↑/↓: 选择 • enter: 确认 • esc: 返回"))
	} else {
		s.WriteString(helpStyle.Render("↑/↓: navigate • enter: launch • q: quit"))
	}

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
