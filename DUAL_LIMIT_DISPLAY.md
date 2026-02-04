# Codex Dual Limit Display - Implementation Details

## Overview

This document explains the sophisticated dual-limit display feature for Codex token usage in Amazing CLI. The implementation shows both 5-hour and weekly limits with gradient-based color coding for enhanced visual clarity.

## Problem Statement

The original Codex CLI output format has evolved to show:
```
5h limit:             [████████████████████] 100% left (resets 03:31 on 5 Feb)
Weekly limit:         [████████████████████] 100% left (resets 16:22 on 10 Feb)
```

This required updating the parser to:
1. Handle both "% left" and "% used" formats
2. Parse absolute reset times (HH:MM on DD MMM)
3. Display BOTH limits simultaneously
4. Provide sophisticated visual styling

## Visual Examples

### Scenario 1: Fresh Limits (0% used = 100% left)
```
▶ ◉   codex               5h:░░░░░░░░░░  0%🟢  Wk:░░░░░░░░░░  0%🟢
```

### Scenario 2: Moderate 5h Usage, Low Weekly
```
▶ ◉   codex               5h:████░░░░░░ 45%🔵  Wk:░░░░░░░░░░  8%🟢
```

### Scenario 3: High 5h Usage, Moderate Weekly
```
▶ ◉   codex               5h:████████░░ 82%🔴  Wk:██████░░░░ 65%💗
```

## Color Coding System

### 5-Hour Limit Colors
- **0-39%**: Bright Green/Turquoise (#00FF88) - Healthy
- **40-59%**: Bright Cyan (#00D9FF) - Light usage
- **60-79%**: Amber/Orange (#FFB000) - Moderate usage
- **80-100%**: Bright Red (#FF0040) - High usage

### Weekly Limit Colors
- **0-39%**: Turquoise (#00FFD4) - Healthy
- **40-59%**: Purple (#9D00FF) - Light usage
- **60-79%**: Hot Pink (#FF69B4) - Moderate usage
- **80-100%**: Deep Pink (#FF1493) - High usage

The different color palettes help users quickly distinguish between the two limit types at a glance.

## Technical Implementation

### Data Structures

```go
// LimitInfo in codex_usage.go
type LimitInfo struct {
    Percentage int    // 0-100, percentage used
    Display    string // Human-readable display
    ResetTime  string // When the limit resets
}

// LimitDetail in tool.go
type LimitDetail struct {
    Percentage int    // 0-100, percentage used
    Display    string // Human-readable display
    ResetTime  string // When the limit resets
}

// UsageInfo with both limits
type UsageInfo struct {
    // ... existing fields ...
    FiveHourLimit LimitInfo // 5h limit details
    WeeklyLimit   LimitInfo // Weekly limit details
}
```

### Parsing Logic

The parser now handles multiple formats:

1. **Old format**: `5h limit: 45% used (resets in 2h 30m)`
2. **New format**: `5h limit: [████] 100% left (resets 03:31 on 5 Feb)`

Key regex patterns:
```go
usedPattern := regexp.MustCompile(`(\d+(?:\.\d+)?)\s*%\s*used`)
leftPattern := regexp.MustCompile(`(\d+(?:\.\d+)?)\s*%\s*left`)
resetInPattern := regexp.MustCompile(`resets in (.+)`)
resetOnPattern := regexp.MustCompile(`resets (\d{2}:\d{2}) on (\d+\s+\w+)`)
```

The parser automatically converts "% left" to "% used":
```go
if percent, err := strconv.ParseFloat(matches[1], 64); err == nil {
    fiveHourPercent = 100 - int(percent) // Convert left to used
    foundFiveHour = true
}
```

### Display Rendering

The `renderDualLimitBar()` function in `tui.go` creates the sophisticated visual display:

```go
func renderDualLimitBar(balance tool.Balance) string {
    // Renders both 5h and weekly limits
    // Uses gradient colors based on usage percentage
    // Returns: "5h:████░░░░░░ 45%  Wk:██░░░░░░░░ 20%"
}
```

## Benefits

1. **At-a-Glance Status**: Users can immediately see both limits without additional commands
2. **Visual Clarity**: Different colors for different limits and usage levels
3. **Compact Display**: Fits on a single line in the TUI
4. **Backward Compatible**: Still works with old format outputs
5. **Sophisticated Styling**: Professional gradient color scheme

## Testing

Comprehensive tests cover:
- Old format parsing ("% used", relative times)
- New format parsing ("% left", absolute times)
- Both limits being captured independently
- Color mapping for all usage levels
- Edge cases (0%, 100%, decimal percentages)

Run tests:
```bash
go test ./pkg/provider/codex -v
```

## Future Enhancements

Possible improvements:
1. Add hover/tooltip showing reset times
2. Animate progress bars when usage changes
3. Show historical usage trends
4. Add notifications when approaching limits
5. Support for additional limit types if Codex adds more

## Architecture Benefits

The modular design allows:
- Easy extension to other AI tools (Copilot, Claude, etc.)
- Simple addition of new limit types
- Flexible color schemes
- Independent testing of each component
