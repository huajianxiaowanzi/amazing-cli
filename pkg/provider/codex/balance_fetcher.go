// Package codex provides functionality to fetch Codex token usage information.
package codex

import (
	"context"

	"github.com/huajianxiaowanzi/amazing-cli/pkg/tool"
)

// BalanceFetcher implements the provider.BalanceFetcher interface for Codex.
type BalanceFetcher struct {
	usageFetcher *UsageFetcher
}

// NewBalanceFetcher creates a new Codex BalanceFetcher.
func NewBalanceFetcher() *BalanceFetcher {
	return &BalanceFetcher{
		usageFetcher: NewUsageFetcher(),
	}
}

// GetBalance fetches the current Codex balance and converts it to tool.Balance.
func (b *BalanceFetcher) GetBalance(ctx context.Context) *tool.Balance {
	usage := b.usageFetcher.GetUsage(ctx)

	return &tool.Balance{
		Percentage: usage.Percentage,
		Display:    usage.Display,
		Color:      usage.Color,
		FiveHourLimit: tool.LimitDetail{
			Percentage: usage.FiveHourLimit.Percentage,
			Display:    usage.FiveHourLimit.Display,
			ResetTime:  usage.FiveHourLimit.ResetTime,
		},
		WeeklyLimit: tool.LimitDetail{
			Percentage: usage.WeeklyLimit.Percentage,
			Display:    usage.WeeklyLimit.Display,
			ResetTime:  usage.WeeklyLimit.ResetTime,
		},
	}
}
