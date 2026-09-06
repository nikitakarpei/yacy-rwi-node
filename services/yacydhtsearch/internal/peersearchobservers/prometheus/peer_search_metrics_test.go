package prometheus_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	prometheusclient "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	peersearchobserversprometheus "github.com/nikitakarpei/yacy-rwi-node/yacydhtsearch/internal/peersearchobservers/prometheus"
)

func publishedBy(t *testing.T, registry *prometheusclient.Registry) string {
	t.Helper()

	recorder := httptest.NewRecorder()
	promhttp.HandlerFor(registry, promhttp.HandlerOpts{}).ServeHTTP(
		recorder,
		httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/metrics", nil),
	)

	return recorder.Body.String()
}

func TestEveryPeerCallIsCountedUnderItsOutcome(t *testing.T) {
	t.Parallel()

	registry := prometheusclient.NewRegistry()
	metrics := peersearchobserversprometheus.New(registry)

	metrics.PeerAnswered(t.Context(), "http://peer.example", 3)
	metrics.PeerRefused(t.Context(), "http://peer.example", http.StatusServiceUnavailable)
	metrics.PeerUnreachable(t.Context(), "http://peer.example", errors.New("no route"))
	metrics.PeerAnswerUnreadable(t.Context(), "http://peer.example", errors.New("bad row"))

	body := publishedBy(t, registry)
	for _, published := range []string{
		`yacydhtsearch_peer_calls_total{outcome="answered"} 1`,
		`yacydhtsearch_peer_calls_total{outcome="refused"} 1`,
		`yacydhtsearch_peer_calls_total{outcome="unreachable"} 1`,
		`yacydhtsearch_peer_calls_total{outcome="unreadable"} 1`,
	} {
		if !strings.Contains(body, published) {
			t.Fatalf("metrics do not carry %q:\n%s", published, body)
		}
	}
}
