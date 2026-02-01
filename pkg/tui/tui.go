// Package tui provides the terminal user interface using Bubble Tea.
package tui

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
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
	title             string
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
	rand.Seed(time.Now().UnixNano())
	title := `    ___                          _                     ___ 
   /   |  ____ ___  ____ _____  (_)___  ____ _   _____/ (_)
  / /| | / __ ` + "`" + `__ \/ __ ` + "`" + `/_  / / / __ \/ __ ` + "`" + `/  / ___/ / / 
 / ___ |/ / / / / / /_/ / / /_/ / / / / /_/ /  / /__/ / /  
/_/  |_/_/ /_/ /_/\__,_/ /___/_/_/ /_/\__, /   \___/_/_/   
                                     /____/               `
	return Model{
		tools:        registry.List(),
		cursor:       0,
		promptCursor: 0,
		spinner:      spin,
		balance:      config.GetDefaultBalance(),
		title:        renderBlockColorTitle(title, rand.Float64()*360.0),
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
	s.WriteString(m.title)
	s.WriteString("\n\n")

	// Tool list
	maxNameWidth := 0
	for _, t := range m.tools {
		// Calculate width with styles applied to account for padding
		w := lipgloss.Width(normalStyle.Render(t.DisplayName))
		if sw := lipgloss.Width(selectedStyle.Render(t.DisplayName)); sw > w {
			w = sw
		}
		if w > maxNameWidth {
			maxNameWidth = w
		}
	}
	const tokenGap = 20
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
		toolNameWidth := lipgloss.Width(toolName)
		balanceBar := renderInlineBalanceBar(m.balance)
		// Calculate padding to align all token bars: (maxNameWidth - currentNameWidth) + fixedGap
		padding := maxNameWidth - toolNameWidth + tokenGap
		s.WriteString(fmt.Sprintf("%s%s %s%s%s\n", cursor, statusIcon, toolName, strings.Repeat(" ", padding), balanceBar))

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

func renderBlockColorTitle(text string, hueOffset float64) string {
	lines := strings.Split(text, "\n")
	height := len(lines)
	maxWidth := 0
	for _, line := range lines {
		if len(line) > maxWidth {
			maxWidth = len(line)
		}
	}
	if maxWidth == 0 || height == 0 {
		return ""
	}

	grid := make([][]rune, height)
	for i, line := range lines {
		row := []rune(line)
		if len(row) < maxWidth {
			padding := make([]rune, maxWidth-len(row))
			for j := range padding {
				padding[j] = ' '
			}
			row = append(row, padding...)
		}
		grid[i] = row
	}

	occupied := make([]bool, maxWidth)
	for c := 0; c < maxWidth; c++ {
		for r := 0; r < height; r++ {
			if grid[r][c] != ' ' {
				occupied[c] = true
				break
			}
		}
	}

	letterIndex := make([]int, maxWidth)
	for i := range letterIndex {
		letterIndex[i] = -1
	}
	currentLetter := 0
	inLetter := false
	for c := 0; c < maxWidth; c++ {
		if occupied[c] {
			if !inLetter {
				inLetter = true
				currentLetter++
			}
			letterIndex[c] = currentLetter - 1
		} else {
			inLetter = false
		}
	}
	totalLetters := currentLetter

	colors := make([]lipgloss.Style, totalLetters)
	hue := hueOffset
	for i := 0; i < totalLetters; i++ {
		rv, gv, bv := hslToRGB(hue, 0.85, 0.55)
		color := lipgloss.Color(fmt.Sprintf("#%02x%02x%02x", rv, gv, bv))
		colors[i] = lipgloss.NewStyle().Foreground(color)
		hue += 35.0
		if hue >= 360.0 {
			hue -= 360.0
		}
	}

	var b strings.Builder
	for r := 0; r < height; r++ {
		for c := 0; c < maxWidth; c++ {
			ch := grid[r][c]
			if ch == ' ' {
				b.WriteRune(ch)
				continue
			}
			idx := letterIndex[c]
			if idx >= 0 && idx < totalLetters {
				b.WriteString(colors[idx].Render(string(ch)))
			} else {
				b.WriteRune(ch)
			}
		}
		if r < height-1 {
			b.WriteRune('\n')
		}
	}
	return b.String()
}

func hslToRGB(h, s, l float64) (uint8, uint8, uint8) {
	h = math.Mod(h, 360.0) / 360.0
	c := (1 - math.Abs(2*l-1)) * s
	x := c * (1 - math.Abs(math.Mod(h*6, 2)-1))
	m := l - c/2

	var r, g, b float64
	switch {
	case 0 <= h && h < 1.0/6.0:
		r, g, b = c, x, 0
	case 1.0/6.0 <= h && h < 2.0/6.0:
		r, g, b = x, c, 0
	case 2.0/6.0 <= h && h < 3.0/6.0:
		r, g, b = 0, c, x
	case 3.0/6.0 <= h && h < 4.0/6.0:
		r, g, b = 0, x, c
	case 4.0/6.0 <= h && h < 5.0/6.0:
		r, g, b = x, 0, c
	default:
		r, g, b = c, 0, x
	}

	r = (r + m) * 255
	g = (g + m) * 255
	b = (b + m) * 255
	return uint8(r + 0.5), uint8(g + 0.5), uint8(b + 0.5)
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
