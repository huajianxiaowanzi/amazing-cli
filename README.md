# 🚀 Amazing CLI

> **One Command. Multiple AI Tools. Zero Hassle.**

A beautiful, lightning-fast TUI launcher for your favorite AI agent command-line tools.

## ✨ Features

- **🎯 One Command to Rule Them All**: Just type `amazing` and you're in
- **⚡ Instant Launch**: Select and execute AI tools with arrow keys + Enter
- **🎨 Gorgeous TUI**: Built with Bubble Tea for a modern, responsive interface
- **📊 Token Balance Tracking**: Ready-to-extend placeholder for monitoring your AI credits
- **🔧 Tool-Agnostic**: Works with claude, copilot, codex, and more
- **🛡️ Graceful Fallback**: Friendly messages when tools aren't installed

## 🎬 Quick Start

```bash
# Install
go install github.com/huajianxiaowanzi/amazing-cli@latest

# Run
amazing
```

Or run directly from source:

```bash
git clone https://github.com/huajianxiaowanzi/amazing-cli.git
cd amazing-cli
go run main.go
```

## 🎮 Usage

1. Launch the TUI: `amazing`
2. Use ↑/↓ arrow keys to navigate
3. Press Enter to launch the selected AI tool
4. Press q to quit

## 🛠️ Supported Tools

- **claude** - Claude Code by Anthropic
- **copilot** - GitHub Copilot CLI
- **codex** - OpenAI's Codex

*Easy to extend with more tools!*

## 📦 Project Structure

```
amazing-cli/
├── main.go              # Entry point
├── go.mod               # Go module definition
├── pkg/
│   ├── config/          # Tool configurations
│   ├── tool/            # Tool interface and registry
│   └── tui/             # Bubble Tea TUI implementation
└── README.md
```

## 🔌 Extending

### Adding a New Tool

```go
// In pkg/config/config.go
tools.Register(&tool.Tool{
    Name:        "your-tool",
    DisplayName: "Your Amazing Tool",
    Command:     "your-tool",
    Description: "Description of your tool",
})
```

### Implementing Token Balance

The token balance system is designed with a clean interface for easy extension:

```go
// Future implementation
type BalanceProvider interface {
    GetBalance(tool string) (Balance, error)
}
```

## 🏗️ Architecture

- **Modular Design**: Clean separation between config, tool management, and UI
- **Interface-Driven**: Easy to mock and test
- **Extensible**: Add new tools or features without touching core logic
- **Type-Safe**: Leverages Go's type system for reliability

## 📝 License

MIT

## 🤝 Contributing

Contributions welcome! Feel free to open issues or submit PRs.

---

Made with ❤️ and ☕
