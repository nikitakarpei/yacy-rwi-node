// Package networksearch ranks what the peers of the configured network hold for
// one query, inside one whole-query time budget.
package networksearch

import (
	"context"
	"time"

	"github.com/nikitakarpei/yacy-rwi-node/yacydhtsearch/internal/peerdirectory"
	"github.com/nikitakarpei/yacy-rwi-node/yacydhtsearch/internal/peersearch"
	"github.com/nikitakarpei/yacy-rwi-node/yacydhtsearch/internal/searchquery"
	"github.com/nikitakarpei/yacy-rwi-node/yacydhtsearch/internal/searchresult"
	"github.com/nikitakarpei/yacy-rwi-node/yacymodel"
	"github.com/nikitakarpei/yacy-rwi-node/yacyproto"
)

type PeerSelection interface {
	PeersFor(
		ctx context.Context,
		query searchquery.Query,
		askable []peerdirectory.AskablePeer,
	) []peerdirectory.AskablePeer
}

type NetworkSearchObserver interface {
	NetworkSearchPerformed(ctx context.Context, asked, answered, items int, spent time.Duration)
	NetworkSearchFoundNoAskablePeers(ctx context.Context)
}

type Network struct {
	networkName string
	directory   *peerdirectory.Directory
	selection   PeerSelection
	peers       peersearch.Peers
	budget      time.Duration
	peerBudget  time.Duration
	peerResults int
	records     int
	partitions  yacymodel.DHTRingPartitions
	observer    NetworkSearchObserver
}

//nolint:revive // argument-limit: nine explicit, independently-meaningful collaborators
func New(
	networkName string,
	directory *peerdirectory.Directory,
	selection PeerSelection,
	peers peersearch.Peers,
	budget, peerBudget time.Duration,
	peerResults, records int,
	partitions yacymodel.DHTRingPartitions,
	observer NetworkSearchObserver,
) Network {
	return Network{
		networkName: networkName,
		directory:   directory,
		selection:   selection,
		peers:       peers,
		budget:      budget,
		peerBudget:  peerBudget,
		peerResults: peerResults,
		records:     records,
		partitions:  partitions,
		observer:    observer,
	}
}

func (n Network) Search(ctx context.Context, query searchquery.Query) searchresult.Ranking {
	ctx, spent := context.WithTimeout(ctx, n.budget)
	defer spent()
	startedAt := time.Now()

	asked := n.selection.PeersFor(ctx, query, n.directory.AskablePeers(ctx))
	if len(asked) == 0 {
		n.observer.NetworkSearchFoundNoAskablePeers(ctx)

		return searchresult.Ranking{}
	}
	n.directory.NoteAsked(ctx, asked)

	answers := n.peers.Ask(ctx, asked, n.requestFor(query))
	ranking := n.rankingOf(answers)
	n.observer.NetworkSearchPerformed(
		ctx, len(asked), len(answers), len(ranking.Items), time.Since(startedAt),
	)

	return ranking
}

func (n Network) requestFor(query searchquery.Query) yacyproto.SearchRequest {
	return yacyproto.SearchRequest{
		NetworkName: n.networkName,
		Query:       query.TermHashes(),
		Exclude:     query.ExclusionHashes(),
		Count:       n.peerResults,
		Time:        int(n.peerBudget.Milliseconds()),
		Partitions:  int(n.partitions),
		ContentDom:  yacyproto.ContentDomainText,
		Language:    query.Language,
	}
}

func (n Network) rankingOf(answers []peersearch.Answer) searchresult.Ranking {
	items := make([][]searchresult.Item, 0, len(answers))
	for _, answer := range answers {
		items = append(items, answer.Items)
	}

	return searchresult.RankingFrom(items, n.records)
}

type NetworkSearchObservers []NetworkSearchObserver

func (observers NetworkSearchObservers) NetworkSearchPerformed(
	ctx context.Context,
	asked, answered, items int,
	spent time.Duration,
) {
	for _, observer := range observers {
		observer.NetworkSearchPerformed(ctx, asked, answered, items, spent)
	}
}

func (observers NetworkSearchObservers) NetworkSearchFoundNoAskablePeers(ctx context.Context) {
	for _, observer := range observers {
		observer.NetworkSearchFoundNoAskablePeers(ctx)
	}
}
