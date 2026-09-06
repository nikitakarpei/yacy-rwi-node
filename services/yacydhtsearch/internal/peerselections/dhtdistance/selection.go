// Package dhtdistance picks the peers a query goes to by how near they sit to
// the postings of its terms on the DHT ring.
package dhtdistance

import (
	"cmp"
	"context"
	"slices"

	"github.com/nikitakarpei/yacy-rwi-node/yacydhtsearch/internal/peerdirectory"
	"github.com/nikitakarpei/yacy-rwi-node/yacydhtsearch/internal/searchquery"
	"github.com/nikitakarpei/yacy-rwi-node/yacymodel"
)

type DHTDistanceObserver interface {
	PeersSelected(ctx context.Context, ringFractions []float64)
}

type Selection struct {
	partitions yacymodel.DHTRingPartitions
	redundancy int
	observer   DHTDistanceObserver
}

func New(
	partitions yacymodel.DHTRingPartitions,
	redundancy int,
	observer DHTDistanceObserver,
) Selection {
	return Selection{partitions: partitions, redundancy: redundancy, observer: observer}
}

func (s Selection) PeersFor(
	ctx context.Context,
	query searchquery.Query,
	askable []peerdirectory.AskablePeer,
) []peerdirectory.AskablePeer {
	chosen := make([]peerdirectory.AskablePeer, 0, len(askable))
	ringFractions := make([]float64, 0, len(askable))
	taken := map[yacymodel.Hash]struct{}{}
	for _, term := range query.TermHashes() {
		for partition := range uint(s.partitions) {
			position := yacymodel.DHTRingPositionOfWordInPartition(term, partition, s.partitions)
			for _, peer := range s.nearest(askable, position) {
				if _, seen := taken[peer.Hash]; seen {
					continue
				}
				taken[peer.Hash] = struct{}{}
				chosen = append(chosen, peer)
				ringFractions = append(ringFractions, ringFractionTo(position, peer))
			}
		}
	}
	s.observer.PeersSelected(ctx, ringFractions)

	return chosen
}

func ringFractionTo(
	position yacymodel.DHTRingPosition,
	peer peerdirectory.AskablePeer,
) float64 {
	return position.DistanceTo(yacymodel.DHTRingPositionOf(peer.Hash)).FractionOfDHTRing()
}

func (s Selection) nearest(
	askable []peerdirectory.AskablePeer,
	position yacymodel.DHTRingPosition,
) []peerdirectory.AskablePeer {
	ranked := slices.SortedFunc(
		slices.Values(askable),
		func(a, b peerdirectory.AskablePeer) int {
			return cmp.Compare(
				position.DistanceTo(yacymodel.DHTRingPositionOf(a.Hash)),
				position.DistanceTo(yacymodel.DHTRingPositionOf(b.Hash)),
			)
		},
	)

	return ranked[:min(s.redundancy, len(ranked))]
}

type DHTDistanceObservers []DHTDistanceObserver

func (observers DHTDistanceObservers) PeersSelected(
	ctx context.Context,
	ringFractions []float64,
) {
	for _, observer := range observers {
		observer.PeersSelected(ctx, ringFractions)
	}
}
