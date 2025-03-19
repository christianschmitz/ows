package network

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/fxamacker/cbor/v2"

	"ows/ledger"
)

const MaxRecentGossips = 100

// TODO: include events and health status
type Gossip struct {
	NodeID  ledger.NodeID
	Head    ledger.ChangeSetID
	Changes []ledger.ChangeSet
}

type encodeableGossip struct {
	NodeID  []byte                       `cbor:"0,keyasint"`
	Head    []byte                       `cbor:"1,keyasint"`
	Changes []ledger.EncodeableChangeSet `cbor:"2,keyasint,omitempty"`
}

type gossipHandler struct {
	kp        *ledger.KeyPair
	callbacks Callbacks

	mutex  sync.Mutex
	recent [][]byte // list of hashes of recent gossips
}

func ServeGossip(port ledger.Port, kp *ledger.KeyPair, callbacks Callbacks) {
	cert, err := makeTLSCertificate(*kp)
	if err != nil {
		panic(err)
	}

	tlsConf := makeServerTLSConfig(cert, func(k ledger.PublicKey) bool {
		l := callbacks.Ledger()

		if _, ok := l.Snapshot.Nodes[k.NodeID()]; ok {
			return true
		} else {
			return false
		}
	})

	s := &http.Server{
		Addr: fmt.Sprintf(":%d", port),
		Handler: &gossipHandler{
			kp:        kp,
			callbacks: callbacks,
		},
		TLSConfig:      tlsConf,
		ReadTimeout:    20 * time.Second,
		WriteTimeout:   20 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	if err := s.ListenAndServeTLS("", ""); err != nil {
		panic(err)
	}
}

func (h *gossipHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "PUT":
		switch r.URL.Path {
		case "/":
			h.servePut(w, r)
		default:
			http.Error(w, fmt.Sprintf("invalid gossip PUT path %s", r.URL.Path), 404)
		}
	default:
		http.Error(w, fmt.Sprintf("invalid gossip HTTP method %s", r.Method), 404)
	}
}

func (h *gossipHandler) isRecentDuplicate(bodyBytes []byte) bool {
	if h.recent == nil {
		h.mutex.Lock()
		h.recent = make([][]byte, 0, 100)
		h.mutex.Unlock()
		return false
	}

	hash := ledger.DigestShort(bodyBytes)

	for _, bs := range h.recent {
		if bytes.Equal(hash, bs) {
			return true
		}
	}

	h.mutex.Lock()
	h.recent = append(h.recent, hash)

	n := len(h.recent)
	if n > MaxRecentGossips {
		h.recent = h.recent[n-MaxRecentGossips:]
	}
	h.mutex.Unlock()

	return false
}

func (h *gossipHandler) servePut(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("invalid request body (%v)", err), 400)
		return
	}

	if h.isRecentDuplicate(body) {
		fmt.Fprintf(w, "")
		return
	}

	l := h.callbacks.Ledger()
	v := l.Snapshot.Version

	g, err := DecodeGossip(body, v)
	if err != nil {
		http.Error(w, fmt.Sprintf("invalid gossip format (%v)", err), 400)
	}

	if g.Head != l.Head() {
		if len(g.Changes) == 0 || g.Changes[len(g.Changes)-1].ID() != g.Head {
			// TODO: fetch changes from API instead
			http.Error(w, fmt.Sprintf("gossip doesn't include necessary changes, aboting"), 400)
			return
		}

		for i, cs := range g.Changes {
			if cs.Prev == l.Head() {
				for j := i; j < len(g.Changes); j++ {
					if err := h.callbacks.AppendChangeSet(&(g.Changes[j])); err != nil {
						http.Error(w, fmt.Sprintf("unable to apply new change set %d (%v)", j, err), 400)
						return
					}
				}

				break
			}
		}
	}

	// spread gossip to other nodes
	gc := NewGossipClient(h.kp, h.callbacks)
	gc.Notify(g)

	fmt.Fprintf(w, "")
}

func (g *Gossip) Encode() []byte {
	_, nodeIDBytes, err := ledger.DecodeBech32(string(g.NodeID))
	if err != nil {
		panic(err)
	}

	_, headBytes, err := ledger.DecodeBech32(string(g.Head))
	if err != nil {
		panic(err)
	}

	ecs := make([]ledger.EncodeableChangeSet, len(g.Changes))

	for i, cs := range g.Changes {
		ecs[i] = ledger.NewEncodeableChangeSet(&cs)
	}

	eg := encodeableGossip{
		NodeID:  nodeIDBytes,
		Head:    headBytes,
		Changes: ecs,
	}

	bs, err := cbor.Marshal(eg)
	if err != nil {
		panic(err)
	}

	return bs
}

func DecodeGossip(bs []byte, v ledger.LedgerVersion) (*Gossip, error) {
	eg := new(encodeableGossip)

	err := cbor.Unmarshal(bs, eg)
	if err != nil {
		return nil, err
	}

	nodeID := ledger.EncodeBech32(ledger.NodeIDPrefix, eg.NodeID)
	head := ledger.EncodeBech32(ledger.ChangeSetIDPrefix, eg.Head)

	changes := make([]ledger.ChangeSet, len(eg.Changes))

	for i, ecs := range eg.Changes {
		cs, err := ecs.ChangeSet(v)
		if err != nil {
			return nil, err
		}

		changes[i] = *cs
	}

	return &Gossip{
		NodeID:  ledger.NodeID(nodeID),
		Head:    ledger.ChangeSetID(head),
		Changes: changes,
	}, nil
}
