//go:build !windows

package codex

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/creack/pty"
)

func runCodexStatus(ctx context.Context, codexPath string) (string, error) {
	// Run codex without restrictions to get full /status output
	cmd := exec.CommandContext(ctx, codexPath)
	// Set environment variables to make codex think it's in a real terminal
	cmd.Env = append(os.Environ(), 
		"TERM=xterm-256color",
		"COLORTERM=truecolor",
		"LINES=60",
		"COLUMNS=160",
	)

	// Set a larger terminal size to ensure full /status output is displayed
	winSize := &pty.Winsize{
		Rows: 60,
		Cols: 160,
		X:    0,
		Y:    0,
	}

	ptmx, err := pty.StartWithSize(cmd, winSize)
	if err != nil {
		return "", fmt.Errorf("failed to start codex with PTY: %w", err)
	}
	defer ptmx.Close()

	var buf bytes.Buffer
	tmp := make([]byte, 8192)
	start := time.Now()
	sentStatus := false
	readyForStatus := false
	statusSentTime := time.Time{}

	// Read output and wait for the prompt before sending /status
	for {
		if time.Since(start) > time.Duration(maxWaitForOutputMs)*time.Millisecond {
			break
		}

		_ = ptmx.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
		n, err := ptmx.Read(tmp)
		if n > 0 {
			chunk := tmp[:n]
			buf.Write(chunk)
			
			// Respond to terminal queries
			if bytes.Contains(chunk, []byte("\x1b[6n")) {
				// Report cursor position
				_, _ = ptmx.Write([]byte("\x1b[30;1R"))
			}
			if bytes.Contains(chunk, []byte("\x1b[c")) || bytes.Contains(chunk, []byte("\x1b[>")) {
				// Report as VT100 compatible terminal with advanced features
				_, _ = ptmx.Write([]byte("\x1b[?62;1;2;6;7;8;9;15;18;21;22c"))
			}
			if bytes.Contains(chunk, []byte("\x1b]10;?")) {
				_, _ = ptmx.Write([]byte("\x1b]10;rgb:ffff/ffff/ffff\x1b\\"))
			}
			if bytes.Contains(chunk, []byte("\x1b]11;?")) {
				_, _ = ptmx.Write([]byte("\x1b]11;rgb:0000/0000/0000\x1b\\"))
			}
			
			// Check if codex is ready (shows prompt with ›)
			cleanOutput := stripANSICodes(buf.String())
			if !readyForStatus && strings.Contains(cleanOutput, "›") && strings.Contains(cleanOutput, "context left") {
				readyForStatus = true
			}
			
			// Send /status once codex is ready
			if readyForStatus && !sentStatus {
				time.Sleep(800 * time.Millisecond)
				// Send /status command and press Enter to trigger the dialog
				if _, err := ptmx.Write([]byte("/status\n")); err != nil {
					return "", fmt.Errorf("failed to send /status command: %w", err)
				}
				sentStatus = true
				statusSentTime = time.Now()
			}
			
			// Check if we got the status output (contains limit info)
			if sentStatus {
				cleanOutput = stripANSICodes(buf.String())
				if strings.Contains(cleanOutput, "5h limit") || strings.Contains(cleanOutput, "Weekly limit") {
					// Give more time to capture complete output
					time.Sleep(500 * time.Millisecond)
					for i := 0; i < 5; i++ {
						_ = ptmx.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
						if n, err := ptmx.Read(tmp); n > 0 && err == nil {
							buf.Write(tmp[:n])
						}
					}
					break
				}
				// Wait at least 5 seconds after sending /status before giving up
				if time.Since(statusSentTime) > 5*time.Second {
					break
				}
			}
		}

		if err != nil {
			if isTimeoutErr(err) {
				continue
			}
			if !errors.Is(err, context.Canceled) {
				break
			}
		}
	}

	if cmd.Process != nil {
		_ = cmd.Process.Kill()
	}

	out := buf.String()
	if out == "" {
		return "", fmt.Errorf("no output from codex /status")
	}
	return out, nil
}

func isTimeoutErr(err error) bool {
	type timeout interface {
		Timeout() bool
	}
	if te, ok := err.(timeout); ok && te.Timeout() {
		return true
	}
	return errors.Is(err, os.ErrDeadlineExceeded)
}
