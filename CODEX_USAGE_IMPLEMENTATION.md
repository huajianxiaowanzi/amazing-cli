# Codex Token Usage Implementation - Visual Comparison

## Before (Original Implementation)
All tools showed fixed 100% token usage:

```
    ___                          _                     ___ 
   /   |  ____ ___  ____ _____  (_)___  ____ _   _____/ (_)
  / /| | / __ `__ \/ __ `/_  / / / __ \/ __ `/  / ___/ / / 
 / ___ |/ / / / / / /_/ / / /_/ / / / / /_/ /  / /__/ / /  
/_/  |_/_/ /_/ /_/\__,_/ /___/_/_/ /_/\__, /   \___/_/_/   
                                     /____/                

▶ ◉   copilot                          Token: 100% ███████████████
  ◉   opencode                         Token: 100% ███████████████
  ◉   codex                            Token: 100% ███████████████
  ○   claude code                      Token: 100% ███████████████
  ○   kimi                             Token: 100% ███████████████

↑/↓: navigate • enter: launch • q: quit
```

## After (New Implementation with Real Codex Data)
Now shows real token usage for Codex (and can be extended for other tools):

```
    ___                          _                     ___ 
   /   |  ____ ___  ____ _____  (_)___  ____ _   _____/ (_)
  / /| | / __ `__ \/ __ `/_  / / / __ \/ __ `/  / ___/ / / 
 / ___ |/ / / / / / /_/ / / /_/ / / / / /_/ /  / /__/ / /  
/_/  |_/_/ /_/ /_/\__,_/ /___/_/_/ /_/\__, /   \___/_/_/   
                                     /____/                

  ◉   copilot                          Token: 100% ███████████████
  ◉   opencode                         Token: 100% ███████████████
▶ ◉   codex                            Token: 45% (2h 30m) ██████░░░░░░░░░
  ○   claude code                      Token: 100% ███████████████
  ○   kimi                             Token: 100% ███████████████

↑/↓: navigate • enter: launch • q: quit
```

### Key Changes:
1. **Codex shows real usage**: `45%` instead of fixed `100%`
2. **Visual progress bar**: Partially filled bar `██████░░░░░░░░░`
3. **Reset time display**: Shows `(2h 30m)` when available
4. **Color coding**: 
   - 🟢 Green (0-59%): Healthy usage
   - 🟡 Yellow (60-79%): Moderate usage
   - 🔴 Red (80-100%): High usage

### Examples with Different Usage Levels:

#### Low Usage (Green)
```
▶ ◉   codex                            Token: 25% ███░░░░░░░░░░░░
```

#### Medium Usage (Yellow)
```
▶ ◉   codex                            Token: 65% █████████░░░░░░
```

#### High Usage (Red)
```
▶ ◉   codex                            Token: 85% (1h) ████████████░░░
```

## How It Works

The implementation uses multiple strategies to fetch Codex token usage:

1. **OAuth API** (Primary): Reads from `~/.codex/auth.json` if available
2. **CLI PTY** (Fallback): Runs `codex /status` command and parses output
3. **Cache** (Performance): Caches results for 5 minutes to avoid excessive API calls
4. **Default** (Graceful Degradation): Falls back to 100% if all strategies fail

### File Structure
```
pkg/
├── provider/
│   ├── provider.go              # BalanceFetcher interface
│   └── codex/
│       ├── codex_usage.go       # Core usage fetching logic
│       ├── codex_usage_test.go  # Unit tests for parsing
│       └── balance_fetcher.go   # Adapter to tool.Balance
├── tool/
│   └── tool.go                  # Tool struct with Balance field
└── tui/
    └── tui.go                   # TUI rendering with per-tool balances
```

## Technical Implementation Details

### Codex Status Output Parsing
The implementation parses output from `codex /status` command:

```
5h limit: 45% used (resets in 2h 30m)
Weekly limit: 10% used (resets in 4 days)
Credits: 1,234.56
```

Using regex patterns to extract:
- Usage percentage: `(\d+(?:\.\d+)?)\s*%\s*used`
- Reset time: `resets in (.+)`

### Color Determination Logic
```go
if usedPercent >= 80 {
    color = "red"
} else if usedPercent >= 60 {
    color = "yellow"
} else {
    color = "green"
}
```

### Caching Strategy
- Cache file: `~/.amazing-cli/cache/codex-usage.json`
- Cache TTL: 5 minutes
- Prevents excessive API/CLI calls while keeping data fresh

## Testing

Comprehensive test coverage for:
- ✅ Output parsing with various formats
- ✅ Color mapping based on usage levels
- ✅ Error handling for invalid data
- ✅ Decimal percentage support
- ✅ Multiple limit types (5h, weekly)

Run tests:
```bash
go test ./pkg/provider/codex -v
```

## Future Enhancements

The architecture is designed to be extensible:

1. **Add More Providers**: Easy to add Copilot, Claude, etc.
   ```go
   case "copilot":
       fetcher := copilot.NewBalanceFetcher()
       balance := fetcher.GetBalance(ctx)
   ```

2. **Parallel Fetching**: Use goroutines for concurrent balance fetching
3. **Periodic Refresh**: Update balances in background
4. **Web Dashboard**: Optional web UI integration like CodexBar

## Benefits

1. **Real-time Visibility**: Users can see actual token consumption
2. **Better Resource Management**: Avoid hitting rate limits
3. **Smart Planning**: Know when limits will reset
4. **Multi-tool Support**: Architecture supports multiple AI tools
5. **Robust Fallback**: Gracefully handles failures
6. **Performance**: Caching minimizes overhead
