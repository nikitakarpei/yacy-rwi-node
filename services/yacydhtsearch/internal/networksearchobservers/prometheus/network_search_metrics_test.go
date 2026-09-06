package prometheus_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	prometheusclient "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	networksearchobserversprometheus "github.com/nikitakarpei/yacy-rwi-node/yacydhtsearch/internal/networksearchobservers/prometheus"
)

const queryBudget = 5 * time.Second

func publishedBy(t *testing.T, registry *prometheusclient.Registry) string {
	t.Helper()

	recorder := httptest.NewRecorder()
	promhttp.HandlerFor(registry, promhttp.HandlerOpts{}).ServeHTTP(
		recorder,
		httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/metrics", nil),
	)

	return recorder.Body.String()
}

func TestOneQueryPublishesThePeersItReachedAndWhatItCost(t *testing.T) {
	t.Parallel()

	registry := prometheusclient.NewRegistry()
	metrics := networksearchobserversprometheus.New(registry, queryBudget)

	metrics.NetworkSearchPerformed(t.Context(), 8, 4, 20, 250*time.Millisecond)

	body := publishedBy(t, registry)
	for _, published := range []string{
		"yacydhtsearch_network_searches_performed_total 1",
		"yacydhtsearch_network_search_peers_asked_sum 8",
		"yacydhtsearch_network_search_completeness_ratio_sum 0.5",
		"yacydhtsearch_network_search_duration_seconds_count 1",
	} {
		if !strings.Contains(body, published) {
			t.Fatalf("metrics do not carry %q:\n%s", published, body)
		}
	}
}

func TestAQueryThatReachedNoPeerIsPublishedAsZeroPeersAsked(t *testing.T) {
	t.Parallel()

	registry := prometheusclient.NewRegistry()
	metrics := networksearchobserversprometheus.New(registry, queryBudget)

	metrics.NetworkSearchFoundNoAskablePeers(t.Context())

	body := publishedBy(t, registry)
	if !strings.Contains(body, "yacydhtsearch_network_search_peers_asked_count 1") ||
		!strings.Contains(body, "yacydhtsearch_network_search_peers_asked_sum 0") {
		t.Fatalf("metrics do not carry a query that asked no peer:\n%s", body)
	}
}

func TestAQueryOverTheBudgetIsCountedApartFromOneInsideIt(t *testing.T) {
	t.Parallel()

	registry := prometheusclient.NewRegistry()
	metrics := networksearchobserversprometheus.New(registry, queryBudget)

	metrics.NetworkSearchPerformed(t.Context(), 1, 1, 1, queryBudget-time.Millisecond)
	metrics.NetworkSearchPerformed(t.Context(), 1, 1, 1, queryBudget+time.Millisecond)

	body := publishedBy(t, registry)
	for _, published := range []string{
		`yacydhtsearch_network_search_duration_seconds_bucket{le="5"} 1`,
		`yacydhtsearch_network_search_duration_seconds_bucket{le="6.25"} 2`,
	} {
		if !strings.Contains(body, published) {
			t.Fatalf("metrics do not carry %q:\n%s", published, body)
		}
	}
}
