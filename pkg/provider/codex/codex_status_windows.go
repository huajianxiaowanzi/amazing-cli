//go:build windows

package codex

import (
	"context"
	"fmt"
)

func runCodexStatus(ctx context.Context, codexPath string) (string, error) {
	_ = ctx
	_ = codexPath
	return "", fmt.Errorf("codex /status requires a TTY; no PTY implementation on windows")
}
