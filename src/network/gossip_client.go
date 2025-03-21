package network

import (
	"bytes"
	"fmt"
	"log"
	"net/http"

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
	l := c.callbacks.Ledger()
	s := l.Snapshot

	initiator := g.NodeID
	dst := OneToClosest(l, c.kp.Public.NodeID(), initiator)

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
