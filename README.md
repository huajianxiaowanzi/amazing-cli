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

### Easy Installation (No Go Required!)

> **Note**: Pre-built binaries require at least one release to be published. If you encounter installation errors, the repository may not have releases yet. See alternative installation methods below.

> **Security Note**: The installation scripts download and verify checksums from GitHub releases. If you prefer to review the scripts before running them, you can download them first:
> - [install.sh](https://raw.githubusercontent.com/huajianxiaowanzi/amazing-cli/main/install.sh) for Linux/macOS
> - [install.ps1](https://raw.githubusercontent.com/huajianxiaowanzi/amazing-cli/main/install.ps1) for Windows

**Linux & macOS:**
```bash
curl -fsSL https://raw.githubusercontent.com/huajianxiaowanzi/amazing-cli/main/install.sh | sh
```

**Windows (PowerShell):**
```powershell
irm https://raw.githubusercontent.com/huajianxiaowanzi/amazing-cli/main/install.ps1 | iex
```

**Manual Download:**

Download pre-built binaries from the [Releases](https://github.com/huajianxiaowanzi/amazing-cli/releases) page:
- Windows: `amazing-cli_Windows_x86_64.zip`
- macOS (Intel): `amazing-cli_Darwin_x86_64.tar.gz`
- macOS (Apple Silicon): `amazing-cli_Darwin_arm64.tar.gz`
- Linux: `amazing-cli_Linux_x86_64.tar.gz`

### For Go Developers

```bash
# Install from source
go install github.com/huajianxiaowanzi/amazing-cli@latest

# Or run directly
go run github.com/huajianxiaowanzi/amazing-cli@latest
```

### After Installation

```bash
# Run the CLI
amazing
```

### Installing AI Tools

After installing Amazing CLI, you can install the AI tools from within the application or manually:

#### Automated Installation
1. Run `amazing`
2. Navigate to an uninstalled tool using arrow keys
3. Press Enter to see installation options
4. Follow the on-screen instructions

#### Manual Installation

**Claude Code:**
```bash
# Linux & macOS
curl -fsSL https://claude.ai/install.sh | bash

# Windows (PowerShell)
irm https://claude.ai/install.ps1 | iex
```
Learn more: [Claude Code Documentation](https://docs.anthropic.com/en/docs/claude-code/getting-started)

**GitHub Copilot:**
```bash
# Linux & macOS
curl -fsSL https://gh.io/copilot-install | bash

# Windows (PowerShell)
winget install GitHub.Copilot

# Alternative (all platforms)
npm install -g @github/copilot
```
Learn more: [GitHub Copilot CLI](https://github.com/github/copilot-cli)

**Kimi Code:**
```bash
# Linux & macOS
curl -L https://code.kimi.com/install.sh | bash

# Windows (PowerShell)
irm https://code.kimi.com/install.ps1 | iex
```

**OpenAI Codex:**
```bash
# macOS
brew install codex || npm i -g @openai/codex

# Linux & Windows
npm i -g @openai/codex
```
Learn more: [OpenAI Codex](https://platform.openai.com/docs/guides/code)

**OpenCode:**
```bash
# All platforms
npm install -g opencode-cli
```
Learn more: [OpenCode CLI](https://github.com/opencode/opencode-cli)

## 🎮 Usage

1. Launch the TUI: `amazing`
2. Use ↑/↓ arrow keys to navigate
3. Press Enter to launch the selected AI tool
4. Press q to quit

## 🛠️ Supported Tools

- **claude** - Claude Code by Anthropic
- **copilot** - GitHub Copilot CLI
- **kimi** - Kimi Code by Moonshot
- **codex** - OpenAI's Codex
- **opencode** - OpenCode AI assistant

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

## 🚀 Creating a Release

To publish pre-built binaries, see [RELEASE.md](RELEASE.md) for detailed instructions.

Quick start:

1. Create and push a version tag:
   ```bash
   git tag -a v0.1.0 -m "Release v0.1.0"
   git push origin v0.1.0
   ```

2. GitHub Actions will automatically:
   - Build binaries for all platforms
   - Create a GitHub release
   - Upload pre-built binaries
   
For first-time setup or troubleshooting, see [RELEASE.md](RELEASE.md).

## 🤝 Contributing

Contributions welcome! Feel free to open issues or submit PRs.

---

Made with ❤️ and ☕
