// Package prometheus reports how often a cached ranking answered a query and
// how often the ranking cache failed.
package prometheus

import (
	"context"

	prometheusclient "github.com/prometheus/client_golang/prometheus"

	"github.com/nikitakarpei/yacy-rwi-node/yacydhtsearch/internal/searchquery"
)

const (
	labelSource    = "source"
	sourceCache    = "cache"
	sourceNetwork  = "network"
	labelAction    = "action"
	actionLookup   = "lookup"
	actionStore    = "store"
	itemBucketBase = 5
)

type RankingMetrics struct {
	lookups  *prometheusclient.CounterVec
	failures *prometheusclient.CounterVec
	items    *prometheusclient.HistogramVec
}

func New(registry prometheusclient.Registerer) *RankingMetrics {
	metrics := &RankingMetrics{
		lookups: prometheusclient.NewCounterVec(prometheusclient.CounterOpts{
			Name: "yacydhtsearch_ranking_lookups_total",
			Help: "Rankings answered to a query, by where the ranking came from.",
		}, []string{labelSource}),
		failures: prometheusclient.NewCounterVec(prometheusclient.CounterOpts{
			Name: "yacydhtsearch_ranking_cache_failures_total",
			Help: "Failures against the ranking cache, by action.",
		}, []string{labelAction}),
		items: prometheusclient.NewHistogramVec(prometheusclient.HistogramOpts{
			Name:    "yacydhtsearch_ranking_items",
			Help:    "Items one ranking carries, by where the ranking came from.",
			Buckets: prometheusclient.LinearBuckets(0, itemBucketBase, itemBucketBase*2),
		}, []string{labelSource}),
	}
	registry.MustRegister(metrics.lookups, metrics.failures, metrics.items)

	return metrics
}

func (m *RankingMetrics) CachedRankingAnswered(_ context.Context, _ searchquery.Query, items int) {
	m.lookups.WithLabelValues(sourceCache).Inc()
	m.items.WithLabelValues(sourceCache).Observe(float64(items))
}

func (m *RankingMetrics) NetworkRankingAnswered(_ context.Context, _ searchquery.Query, items int) {
	m.lookups.WithLabelValues(sourceNetwork).Inc()
	m.items.WithLabelValues(sourceNetwork).Observe(float64(items))
}

func (m *RankingMetrics) RankingLookupFailed(context.Context, searchquery.Query, error) {
	m.failures.WithLabelValues(actionLookup).Inc()
}

func (m *RankingMetrics) RankingStoreFailed(context.Context, searchquery.Query, error) {
	m.failures.WithLabelValues(actionStore).Inc()
}
