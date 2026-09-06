// Package prometheus counts calls to peers by outcome, and what each call cost.
package prometheus

import (
	"context"

	prometheusclient "github.com/prometheus/client_golang/prometheus"
)

const (
	labelOutcome           = "outcome"
	outcomePeerAnswered    = "answered"
	outcomePeerRefused     = "refused"
	outcomePeerUnreachable = "unreachable"
	outcomePeerUnreadable  = "unreadable"
)

type PeerSearchMetrics struct {
	peerCalls *prometheusclient.CounterVec
}

func New(registry prometheusclient.Registerer) *PeerSearchMetrics {
	metrics := &PeerSearchMetrics{
		peerCalls: prometheusclient.NewCounterVec(prometheusclient.CounterOpts{
			Name: "yacydhtsearch_peer_calls_total",
			Help: "Calls to peers, by outcome.",
		}, []string{labelOutcome}),
	}
	registry.MustRegister(metrics.peerCalls)

	return metrics
}

func (m *PeerSearchMetrics) PeerAnswered(_ context.Context, _ string, _ int) {
	m.peerCalls.WithLabelValues(outcomePeerAnswered).Inc()
}

func (m *PeerSearchMetrics) PeerRefused(_ context.Context, _ string, _ int) {
	m.peerCalls.WithLabelValues(outcomePeerRefused).Inc()
}

func (m *PeerSearchMetrics) PeerUnreachable(_ context.Context, _ string, _ error) {
	m.peerCalls.WithLabelValues(outcomePeerUnreachable).Inc()
}

func (m *PeerSearchMetrics) PeerAnswerUnreadable(_ context.Context, _ string, _ error) {
	m.peerCalls.WithLabelValues(outcomePeerUnreadable).Inc()
}
