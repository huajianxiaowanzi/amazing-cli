// Package tui provides the terminal user interface using Bubble Tea.
package tui

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
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

// Styles for the TUI - Cyberpunk Theme
var (
	// Cyberpunk Neon Colors
	neonCyan   = lipgloss.Color("#00F5FF")
	neonPink   = lipgloss.Color("#FF00FF")
	neonPurple = lipgloss.Color("#9D00FF")
	neonYellow = lipgloss.Color("#FFFF00")
	neonGreen  = lipgloss.Color("#39FF14")
	neonOrange = lipgloss.Color("#FF9500")
	neonRed    = lipgloss.Color("#FF0040")
	darkBg     = lipgloss.Color("#0D0D0D")
	gridDark   = lipgloss.Color("#1A1A2E")
	gridLine   = lipgloss.Color("#16213E")
	glowWhite  = lipgloss.Color("#E0E0E0")
	mutedText  = lipgloss.Color("#6B7280")

	// Title - 保持彩虹效果
	titleStyle = lipgloss.NewStyle().
			MarginTop(1).
			MarginBottom(2)

	// Selected Item - 赛博朋克霓虹效果
	selectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#000000")).
			Background(neonCyan).
			PaddingLeft(2).
			PaddingRight(2)

	// Normal Item
	normalStyle = lipgloss.NewStyle().
			Foreground(glowWhite).
			PaddingLeft(2).
			PaddingRight(2)

	// Submenu Items - 无背景色，仅用前景色区分，无padding
	submenuStyle = lipgloss.NewStyle().
			Foreground(mutedText)

	submenuSelectedStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(neonCyan)

	// Status Icons - 赛博朋克风格
	installedStyle = lipgloss.NewStyle().
			Foreground(neonGreen).
			Bold(true)

	notInstalledStyle = lipgloss.NewStyle().
				Foreground(neonRed).
				Bold(true)

	// Token Balance Bar
	balanceStyle = lipgloss.NewStyle().
			Foreground(neonCyan).
			Bold(true)

	// Description & Help
	descStyle = lipgloss.NewStyle().
			Foreground(mutedText).
			Italic(true).
			PaddingLeft(2)

	helpStyle = lipgloss.NewStyle().
			Foreground(mutedText).
			MarginTop(2).
			MarginBottom(1)

	// Dialog & Messages
	dialogStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(neonCyan).
			Background(gridDark).
			Padding(1, 2).
			MarginTop(1).
			MarginBottom(1)

	// Status Messages
	successMsgStyle = lipgloss.NewStyle().
			Foreground(neonGreen).
			Bold(true).
			PaddingLeft(2)

	errorMsgStyle = lipgloss.NewStyle().
			Foreground(neonRed).
			Bold(true).
			PaddingLeft(2)

	warningStyle = lipgloss.NewStyle().
			Foreground(neonYellow).
			Bold(true).
			PaddingLeft(2)
)

// Model represents the TUI state.
type Model struct {
	tools             []*tool.Tool
	cursor            int
	promptCursor      int
	spinner           spinner.Model
	selected          string
	title             string
	quitting          bool
	err               error
	showInstallPrompt bool
	installing        bool
	installError      string
	installSuccess    bool
	terminalHeight    int // 终端高度，用于固定底部帮助文本
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
	case tea.WindowSizeMsg:
		// 记录终端高度，用于固定底部帮助文本
		m.terminalHeight = msg.Height
		return m, nil

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
					// Cancel - close prompt
					m.showInstallPrompt = false
					m.installError = ""
					m.installSuccess = false
					return m, nil
				}
				// Install (promptCursor == 1)
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
			// User selected a tool - 需要先排序获取正确的工具
			sortedTools := m.getSortedTools()
			selectedTool := sortedTools[m.cursor]

			// Check if tool is installed
			if !selectedTool.IsInstalled() {
				// Show install prompt
				m.showInstallPrompt = true
				m.promptCursor = 0
				return m, nil
			}

			// Tool is installed, update last used time and proceed to launch
			selectedTool.LastUsed = time.Now()
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

	// Tool list - 按安装状态分组，已安装的按LRU排序
	sortedTools := m.getSortedTools()

	maxNameWidth := 0
	for _, t := range sortedTools {
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
	for i, t := range sortedTools {
		isSelected := m.cursor == i
		style := normalStyle

		// Cursor indicator
		var cursor string
		if isSelected {
			style = selectedStyle
			cursor = lipgloss.NewStyle().
				Foreground(neonCyan).
				Bold(true).
				Render("▶ ")
		} else {
			cursor = lipgloss.NewStyle().
				Foreground(gridLine).
				Render("  ")
		}

		// Check if tool is installed
		var statusIcon string
		if t.IsInstalled() {
			statusIcon = installedStyle.Render("◉")
		} else {
			statusIcon = notInstalledStyle.Render("○")
		}

		// Render tool item with inline token balance
		toolName := style.Render(t.DisplayName)
		toolNameWidth := lipgloss.Width(toolName)
		
		// Get balance for this tool
		balance := getToolBalance(t)
		balanceBar := renderInlineBalanceBar(balance)
		
		// Calculate padding to align all token bars: (maxNameWidth - currentNameWidth) + fixedGap
		padding := maxNameWidth - toolNameWidth + tokenGap
		s.WriteString(fmt.Sprintf("%s%s %s%s%s\n", cursor, statusIcon, toolName, strings.Repeat(" ", padding), balanceBar))

		// Inline install options when tool is not installed and selected - 两行箭头显示
		if m.showInstallPrompt && m.cursor == i && !t.IsInstalled() {
			cancelLabel := "Cancel"
			installLabel := "Install"
			if !t.HasInstallCommand() {
				installLabel = "Install (N/A)"
			}

			// Cancel 行 - 选中时显示»，未选中时显示空格
			if m.promptCursor == 0 {
				s.WriteString(fmt.Sprintf("      %s %s\n", submenuSelectedStyle.Render("»"), submenuSelectedStyle.Render(cancelLabel)))
			} else {
				s.WriteString(fmt.Sprintf("       %s\n", submenuStyle.Render(cancelLabel)))
			}

			// Install 行 - 选中时显示»，未选中时显示空格
			if m.promptCursor == 1 {
				s.WriteString(fmt.Sprintf("      %s %s\n", submenuSelectedStyle.Render("»"), submenuSelectedStyle.Render(installLabel)))
			} else {
				s.WriteString(fmt.Sprintf("       %s\n", submenuStyle.Render(installLabel)))
			}
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
		s.WriteString(helpStyle.Render("↑/↓: select • enter: confirm • esc: cancel"))
	} else {
		s.WriteString(helpStyle.Render("↑/↓: navigate • enter: launch • q: quit"))
	}

	return s.String()
}

// GetSelected returns the name of the selected tool, if any.
func (m Model) GetSelected() string {
	return m.selected
}

// getSortedTools returns tools sorted by installation status and LRU (最近使用的在前)
func (m Model) getSortedTools() []*tool.Tool {
	sorted := make([]*tool.Tool, len(m.tools))
	copy(sorted, m.tools)

	sort.SliceStable(sorted, func(i, j int) bool {
		installedI := sorted[i].IsInstalled()
		installedJ := sorted[j].IsInstalled()

		// 如果安装状态不同，已安装的排在前面
		if installedI != installedJ {
			return installedI && !installedJ
		}

		// 如果都已安装，按最后使用时间降序排序（最近使用的在前）
		if installedI && installedJ {
			return sorted[i].LastUsed.After(sorted[j].LastUsed)
		}

		// 都未安装，保持原有顺序
		return false
	})

	return sorted
}

// getToolBalance returns the balance for a given tool.
// If the tool's balance hasn't been fetched yet, it returns a default balance.
func getToolBalance(t *tool.Tool) tool.Balance {
	if t.Balance != nil {
		return *t.Balance
	}
	// Return default balance if not fetched using the conversion method
	return config.GetDefaultBalance().ToToolBalance()
}

// renderInlineBalanceBar creates a compact visual representation of the token balance.
// For Codex, it shows both 5h and weekly limits with sophisticated styling.
func renderInlineBalanceBar(balance tool.Balance) string {
	// Check if this is Codex with dual limits
	hasBothLimits := balance.FiveHourLimit.Display != "" || balance.WeeklyLimit.Display != ""
	
	if hasBothLimits {
		return renderDualLimitBar(balance)
	}
	
	// Original single limit display
	width := 15
	percentage := balance.Percentage
	if percentage < 0 {
		percentage = 0
	}
	if percentage > 100 {
		percentage = 100
	}

	filled := (width * percentage) / 100
	empty := width - filled

	filledBar := strings.Repeat("█", filled)
	emptyBar := strings.Repeat("░", empty)

	var barColor lipgloss.Color
	switch balance.Color {
	case "green":
		barColor = neonGreen
	case "yellow":
		barColor = neonYellow
	case "red":
		barColor = neonRed
	default:
		barColor = neonGreen
	}

	barStyle := lipgloss.NewStyle().Foreground(barColor)
	emptyStyle := lipgloss.NewStyle().Foreground(gridLine)

	labelStyle := lipgloss.NewStyle().
		Foreground(neonCyan).
		Bold(true)

	label := labelStyle.Render(fmt.Sprintf("Token: %s", balance.Display))
	barStr := barStyle.Render(filledBar) + emptyStyle.Render(emptyBar)

	return fmt.Sprintf("%s %s", label, barStr)
}

// renderDualLimitBar creates a sophisticated dual-limit display for Codex.
func renderDualLimitBar(balance tool.Balance) string {
	barWidth := 10
	
	// Render 5h limit bar
	fiveHourBar := ""
	if balance.FiveHourLimit.Display != "" {
		percentage := balance.FiveHourLimit.Percentage
		if percentage < 0 {
			percentage = 0
		}
		if percentage > 100 {
			percentage = 100
		}
		
		filled := (barWidth * percentage) / 100
		empty := barWidth - filled
		
		// Sophisticated gradient colors for 5h limit
		var barColor lipgloss.Color
		if percentage >= 80 {
			barColor = lipgloss.Color("#FF0040") // Bright red
		} else if percentage >= 60 {
			barColor = lipgloss.Color("#FFB000") // Amber/orange
		} else if percentage >= 40 {
			barColor = lipgloss.Color("#00D9FF") // Bright cyan
		} else {
			barColor = lipgloss.Color("#00FF88") // Bright green
		}
		
		filledStyle := lipgloss.NewStyle().Foreground(barColor).Bold(true)
		emptyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#2A2A3E"))
		
		filledBar := filledStyle.Render(strings.Repeat("█", filled))
		emptyBar := emptyStyle.Render(strings.Repeat("░", empty))
		
		labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#8BE9FD")).Bold(true)
		label := labelStyle.Render("5h")
		
		percentStyle := lipgloss.NewStyle().Foreground(barColor)
		percentStr := percentStyle.Render(fmt.Sprintf("%d%%", percentage))
		
		fiveHourBar = fmt.Sprintf("%s:%s%s %s", label, filledBar, emptyBar, percentStr)
	}
	
	// Render weekly limit bar
	weeklyBar := ""
	if balance.WeeklyLimit.Display != "" {
		percentage := balance.WeeklyLimit.Percentage
		if percentage < 0 {
			percentage = 0
		}
		if percentage > 100 {
			percentage = 100
		}
		
		filled := (barWidth * percentage) / 100
		empty := barWidth - filled
		
		// Sophisticated gradient colors for weekly limit
		var barColor lipgloss.Color
		if percentage >= 80 {
			barColor = lipgloss.Color("#FF1493") // Deep pink
		} else if percentage >= 60 {
			barColor = lipgloss.Color("#FF69B4") // Hot pink
		} else if percentage >= 40 {
			barColor = lipgloss.Color("#9D00FF") // Purple
		} else {
			barColor = lipgloss.Color("#00FFD4") // Turquoise
		}
		
		filledStyle := lipgloss.NewStyle().Foreground(barColor).Bold(true)
		emptyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#2A2A3E"))
		
		filledBar := filledStyle.Render(strings.Repeat("█", filled))
		emptyBar := emptyStyle.Render(strings.Repeat("░", empty))
		
		labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#BD93F9")).Bold(true)
		label := labelStyle.Render("Wk")
		
		percentStyle := lipgloss.NewStyle().Foreground(barColor)
		percentStr := percentStyle.Render(fmt.Sprintf("%d%%", percentage))
		
		weeklyBar = fmt.Sprintf("%s:%s%s %s", label, filledBar, emptyBar, percentStr)
	}
	
	// Combine both bars
	if fiveHourBar != "" && weeklyBar != "" {
		return fmt.Sprintf("%s  %s", fiveHourBar, weeklyBar)
	} else if fiveHourBar != "" {
		return fiveHourBar
	} else if weeklyBar != "" {
		return weeklyBar
	}
	
	// Fallback
	return renderInlineBalanceBar(balance)
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

	// Cyberpunk neon color palette for title
	cyberpunkColors := []string{
		"#00F5FF", // 霓虹青
		"#FF00FF", // 霓虹粉
		"#9D00FF", // 霓虹紫
		"#39FF14", // 霓虹绿
		"#FF9500", // 霓虹橙
		"#FF0040", // 霓虹红
		"#00FFFF", // 青色
		"#FF1493", // 深粉
		"#7FFF00", // 黄绿
		"#FF69B4", // 热粉
	}

	colors := make([]lipgloss.Style, totalLetters)
	for i := 0; i < totalLetters; i++ {
		colorIdx := (i + int(hueOffset/36)) % len(cyberpunkColors)
		colors[i] = lipgloss.NewStyle().
			Foreground(lipgloss.Color(cyberpunkColors[colorIdx])).
			Bold(true)
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
