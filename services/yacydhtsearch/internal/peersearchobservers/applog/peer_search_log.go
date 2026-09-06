// Package applog reports the outcome of each call to a peer to the log.
package applog

import (
	"context"
	"log/slog"
)

const (
	msgPeerAnswered         = "peer answered a search"
	msgPeerRefused          = "peer refused a search"
	msgPeerUnreachable      = "peer could not be reached for a search"
	msgPeerAnswerUnreadable = "peer answered a search unreadably"
)

type PeerSearchLog struct{}

func (PeerSearchLog) PeerAnswered(ctx context.Context, address string, resources int) {
	slog.DebugContext(ctx, msgPeerAnswered,
		slog.String("address", address),
		slog.Int("resources", resources),
	)
}

func (PeerSearchLog) PeerRefused(ctx context.Context, address string, status int) {
	slog.WarnContext(ctx, msgPeerRefused,
		slog.String("address", address),
		slog.Int("status", status),
	)
}

func (PeerSearchLog) PeerUnreachable(ctx context.Context, address string, cause error) {
	slog.WarnContext(ctx, msgPeerUnreachable,
		slog.String("address", address),
		slog.Any("error", cause),
	)
}

func (PeerSearchLog) PeerAnswerUnreadable(ctx context.Context, address string, cause error) {
	slog.WarnContext(ctx, msgPeerAnswerUnreadable,
		slog.String("address", address),
		slog.Any("error", cause),
	)
}
