// Package applog reports what one network search did to the service log.
package applog

import (
	"context"
	"log/slog"
	"time"
)

const (
	msgNetworkSearchPerformed           = "network search performed"
	msgNetworkSearchFoundNoAskablePeers = "network search found no askable peers"
)

type NetworkSearchLog struct{}

func (NetworkSearchLog) NetworkSearchPerformed(
	ctx context.Context,
	asked, answered, items int,
	spent time.Duration,
) {
	slog.DebugContext(ctx, msgNetworkSearchPerformed,
		slog.Int("askedPeers", asked),
		slog.Int("answeredPeers", answered),
		slog.Int("items", items),
		slog.Duration("spent", spent),
	)
}

func (NetworkSearchLog) NetworkSearchFoundNoAskablePeers(ctx context.Context) {
	slog.WarnContext(ctx, msgNetworkSearchFoundNoAskablePeers)
}
