// Package applog reports the outcome of each call to a peer to the log.
package applog

import (
	"context"
	"log/slog"
	"time"
)

const (
	msgPeerAnswered         = "peer answered a search"
	msgPeerRefused          = "peer refused a search"
	msgPeerUnreachable      = "peer could not be reached for a search"
	msgPeerAnswerUnreadable = "peer answered a search unreadably"
)

type PeerSearchLog struct{}

func (PeerSearchLog) PeerAnswered(
	ctx context.Context,
	address string,
	resources int,
	spent time.Duration,
) {
	slog.DebugContext(ctx, msgPeerAnswered,
		slog.String("address", address),
		slog.Int("resources", resources),
		slog.Duration("spent", spent),
	)
}

func (PeerSearchLog) PeerRefused(
	ctx context.Context,
	address string,
	status int,
	spent time.Duration,
) {
	slog.WarnContext(ctx, msgPeerRefused,
		slog.String("address", address),
		slog.Int("status", status),
		slog.Duration("spent", spent),
	)
}

func (PeerSearchLog) PeerUnreachable(
	ctx context.Context,
	address string,
	cause error,
	spent time.Duration,
) {
	slog.WarnContext(ctx, msgPeerUnreachable,
		slog.String("address", address),
		slog.Any("error", cause),
		slog.Duration("spent", spent),
	)
}

func (PeerSearchLog) PeerAnswerUnreadable(
	ctx context.Context,
	address string,
	cause error,
	spent time.Duration,
) {
	slog.WarnContext(ctx, msgPeerAnswerUnreadable,
		slog.String("address", address),
		slog.Any("error", cause),
		slog.Duration("spent", spent),
	)
}
