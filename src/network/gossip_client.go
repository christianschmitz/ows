package network

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"sort"

	"ows/ledger"
)

type GossipClient struct {
	kp         *ledger.KeyPair
	httpClient *http.Client
	callbacks  Callbacks
}

func NewGossipClient(kp *ledger.KeyPair, callbacks Callbacks) *GossipClient {
	cert, err := makeTLSCertificate(*kp)
	if err != nil {
		panic(err)
	}

	tlsConf := makeClientTLSConfig(cert, func(peer ledger.PublicKey) bool {
		l := callbacks.Ledger()

		if _, ok := l.Snapshot.Nodes[peer.NodeID()]; ok {
			return true
		} else {
			return false
		}
	})

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConf,
		},
	}

	return &GossipClient{kp, httpClient, callbacks}
}

func (c *GossipClient) Notify(g *Gossip) {
	initiator := g.NodeID
	dst := c.oneToClosest(initiator)

	l := c.callbacks.Ledger()
	s := l.Snapshot
	bs := g.Encode()

	for _, id := range dst {
		conf, ok := s.Nodes[id]

		if !ok {
			panic("node not found")
		}

		address := conf.Address
		port := conf.GossipPort

		url := fmt.Sprintf("https://%s:%d/", address, port)

		req, err := http.NewRequest("PUT", url, bytes.NewBuffer(bs))
		if err != nil {
			panic("failed to create PUT request")
		}

		if _, err := c.httpClient.Do(req); err != nil {
			log.Printf("failed to gossip to %s (%v)", url, err)
		}
	}

	return
}

// the most basic topology: one-to-all
func (c *GossipClient) oneToAll(initiator ledger.NodeID) []ledger.NodeID {
	currentNodeID := c.kp.Public.NodeID()

	dst := make([]ledger.NodeID, 0)

	l := c.callbacks.Ledger()
	s := l.Snapshot

	for id, _ := range s.Nodes {
		if initiator == currentNodeID && id != initiator {
			dst = append(dst, id)
		}
	}

	return dst
}

const closest = 10
const overlap = 3

// a more robust topology: one to ten closest,
func (c *GossipClient) oneToClosest(initiator ledger.NodeID) []ledger.NodeID {
	l := c.callbacks.Ledger()
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

		allCpy := all[:]

		sort.Slice(allCpy, func(i, j int) bool {
			di := ledger.HammingDistance(string(current), string(allCpy[i]))
			dj := ledger.HammingDistance(string(current), string(allCpy[j]))

			if di == dj {
				return allCpy[i] < allCpy[j]
			} else {
				return di < dj
			}
		})

		for _, other := range allCpy {
			if len(dst) >= closest {
				break
			}

			if rcv, ok := receiveFrom[other]; ok && len(rcv) >= overlap {
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

	return sendTo[c.kp.Public.NodeID()]
}
