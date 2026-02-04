// Package provider defines interfaces for fetching tool-specific balance information.
package provider

import (
	"context"

	"github.com/huajianxiaowanzi/amazing-cli/pkg/tool"
)

// BalanceFetcher is the interface for fetching balance information for a specific tool.
type BalanceFetcher interface {
	// GetBalance fetches the current balance/usage for the tool.
	GetBalance(ctx context.Context) *tool.Balance
}
