package network

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"ows/ledger"
)

// API server handler
type apiHandler struct {
	callbacks Callbacks
}

func ServeAPI(port ledger.Port, kp *ledger.KeyPair, callbacks Callbacks) {
	cert, err := makeTLSCertificate(*kp)
	if err != nil {
		panic(err)
	}

	tlsConf := makeServerTLSConfig(cert, func(k ledger.PublicKey) bool {
		l := callbacks.Ledger()

		if _, ok := l.Snapshot.Nodes[k.NodeID()]; ok {
			return true
		} else if _, ok := l.Snapshot.Users[k.UserID()]; ok {
			return true
		} else {
			return false
		}
	})

	s := &http.Server{
		Addr:           fmt.Sprintf(":%d", port),
		Handler:        &apiHandler{callbacks},
		TLSConfig:      tlsConf,
		ReadTimeout:    20 * time.Second,
		WriteTimeout:   20 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	if err := s.ListenAndServeTLS("", ""); err != nil {
		panic(err)
	}
}

func (h *apiHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO: only show this with a high verbosity level
	log.Printf("api %s %s\n", r.Method, r.URL.Path)

	switch r.Method {
	case "GET":
		switch r.URL.Path {
		case "/":
			h.serveChangeSetIDChain(w, r)
		case "/assets":
			h.serveGetAssetList(w, r)
		case "/head":
			h.serveHead(w, r)
		default:
			if strings.HasPrefix(r.URL.Path, "/assets/") {
				h.serveGetAsset(w, r)
			} else {
				h.serveChangeSet(w, r)
			}
		}
	case "POST":
		switch r.URL.Path {
		case "/":
			h.servePostChangeSet(w, r)
		default:
			http.Error(w, fmt.Sprintf("unhandled POST path %s", r.URL.Path), 404)
		}
	case "PUT":
		switch r.URL.Path {
		case "/assets":
			h.servePutAsset(w, r)
		default:
			http.Error(w, fmt.Sprintf("unhandled PUT path %s", r.URL.Path), 404)
		}
	default:
		http.Error(w, fmt.Sprintf("unsupported node API HTTP method %s", r.Method), 404)
	}
}

func (h *apiHandler) serveGetAsset(w http.ResponseWriter, r *http.Request) {
	id := strings.Trim(r.URL.Path[len("/assets/"):], "/")

	if err := ledger.ValidateID(id, ledger.AssetIDPrefix); err != nil {
		log.Printf("invalid asset id %s (%v)\n", id, err)
		http.Error(w, fmt.Sprintf("invalid asset id %s (%v)", id, err), 400)
		return
	}

	bs, err := h.callbacks.GetAsset(ledger.AssetID(id))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Printf("requested asset %s not found\n", id)
			http.Error(w, fmt.Sprintf("asset %s not found", id), 404)
		} else {
			log.Printf("internal failure in serveGetAsset (%v)\n", err)
			http.Error(w, fmt.Sprintf("%v", err), 500)
		}
	} else {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(bs)
	}
}

func (h *apiHandler) serveGetAssetList(w http.ResponseWriter, r *http.Request) {
	assets := h.callbacks.ListAssets()

	bs, err := json.Marshal(assets)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to create asset list json (%v)", err), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s", string(bs))
}

func (h *apiHandler) serveChangeSet(w http.ResponseWriter, r *http.Request) {
	id := strings.Trim(r.URL.Path, "/")

	if err := ledger.ValidateID(id, ledger.ChangeSetIDPrefix); err != nil {
		http.Error(w, fmt.Sprintf("invalid change set id %s (%v)", id, err), 400)
		return
	}

	cs, ok := h.callbacks.Ledger().FindChange(ledger.ChangeSetID(id))
	if !ok {
		http.Error(w, fmt.Sprintf("invalid path %s", r.URL.Path), 404)
		return
	}

	bs := cs.Encode()

	w.Header().Set("Content-Type", "application/cbor")
	w.Write(bs)
}

func (h *apiHandler) serveChangeSetIDChain(w http.ResponseWriter, r *http.Request) {
	chain := h.callbacks.Ledger().IDChain()

	bs, err := json.Marshal(chain.IDs)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to create ID chain json (%v)", err), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s", string(bs))
}

func (h *apiHandler) serveHead(w http.ResponseWriter, r *http.Request) {
	head := h.callbacks.Ledger().Head()

	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "%s", head)
}

func (h *apiHandler) servePostChangeSet(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("invalid request body (%v)", err), 400)
		return
	}

	cs, err := ledger.DecodeChangeSet(body, h.callbacks.Ledger().Snapshot.Version)
	if err != nil {
		http.Error(w, fmt.Sprintf("invalid change set format (%v)", err), 400)
		return
	}

	if err := h.callbacks.AppendChangeSet(cs); err != nil {
		http.Error(w, fmt.Sprintf("failure while appending change set (%v)", err), 400)
		return
	}

	fmt.Fprintf(w, "")
}

func (h *apiHandler) servePutAsset(w http.ResponseWriter, r *http.Request) {
	isFromNode := h.hasNodeCertificate(r)

	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("invalid request body (%v)", err), 400)
		return
	}

	id, err := h.callbacks.AddAsset(body, isFromNode)
	if err != nil {
		http.Error(w, fmt.Sprintf("%v", err), 500)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "%s", id)
}

func (h *apiHandler) hasNodeCertificate(r *http.Request) bool {
	if r.TLS == nil {
		return false
	}

	if r.TLS.PeerCertificates == nil {
		return false
	}

	if len(r.TLS.PeerCertificates) == 0 {
		return false
	}

	for _, peerCert := range r.TLS.PeerCertificates {
		key, err := extractPeerPublicKey(peerCert)
		if err != nil {
			continue
		}

		if _, ok := h.callbacks.Ledger().Snapshot.Nodes[key.NodeID()]; ok {
			return true
		}
	}

	return false
}
