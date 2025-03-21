package network

import (
	"sort"

	"ows/ledger"
)

const (
	closest            = 10
	TopologyRedundancy = 3
)

// the most basic topology: one-to-all
func OneToAll(l *ledger.Ledger, current, initiator ledger.NodeID) []ledger.NodeID {
	dst := make([]ledger.NodeID, 0)

	s := l.Snapshot

	for id, _ := range s.Nodes {
		if initiator == current && id != initiator {
			dst = append(dst, id)
		}
	}

	return dst
}

// a more robust topology: one to ten closest,
func OneToClosest(l *ledger.Ledger, current, initiator ledger.NodeID) []ledger.NodeID {
	snapshot := l.Snapshot

	// collect all NodeIDs, ignoring initiator
	all := []ledger.NodeID{}
	for id, _ := range snapshot.Nodes {
		if id != initiator {
			all = append(all, id)
		}
	}

	sendTo := map[ledger.NodeID][]ledger.NodeID{}
	receiveFrom := map[ledger.NodeID][]ledger.NodeID{}

	stack := []ledger.NodeID{initiator}

	for len(stack) > 0 {
		current := stack[0]
		stack = stack[1:]

		// if current has already been treated: continue
		if _, ok := sendTo[current]; ok {
			continue
		}

		dst := []ledger.NodeID{}

		allCpy := make([]ledger.NodeID, 0)
		for _, id := range all {
			// don't include the current node in the list of closest nodes
			if id != current {
				allCpy = append(allCpy, id)
			}
		}

		sortNodesByDistanceToTarget(allCpy, string(current))

		for _, other := range allCpy {
			if len(dst) >= closest {
				break
			}

			if rcv, ok := receiveFrom[other]; ok && len(rcv) >= TopologyRedundancy {
				continue
			}

			dst = append(dst, other)

			if rcv, ok := receiveFrom[other]; ok {
				receiveFrom[other] = append(rcv, current)
			} else {
				receiveFrom[other] = []ledger.NodeID{current}
			}
		}

		sendTo[current] = dst

		stack = append(stack, dst...)
	}

	return sendTo[current]
}

func ClosestNodes(l *ledger.Ledger, resourceID string, n int) []ledger.NodeID {
	nodeIDs := l.Snapshot.NodeIDs()

	sortNodesByDistanceToTarget(nodeIDs, resourceID)

	if len(nodeIDs) > n {
		return nodeIDs[0:n]
	} else {
		return nodeIDs
	}
}

func sortNodesByDistanceToTarget(ids []ledger.NodeID, target string) {
	sort.Slice(ids, func(i, j int) bool {
		di := ledger.HammingDistance(target, string(ids[i]))
		dj := ledger.HammingDistance(target, string(ids[j]))

		if di == dj {
			return ids[i] < ids[j]
		} else {
			return di < dj
		}
	})
}
