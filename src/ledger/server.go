package ledger

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const SYNC_PORT = 9000

type syncHandler struct {
	ledger          *Ledger
	resourceManager ResourceManager
}

func ListenAndServeLedger(l *Ledger, rm ResourceManager) {
	s := &http.Server{
		Addr:           ":" + strconv.Itoa(SYNC_PORT),
		Handler:        &syncHandler{l, rm},
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Fatal(s.ListenAndServe())
}

func (h *syncHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		switch r.URL.Path {
		case "/":
			// return a list of all the heads
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, "%s", h.getLedgerChangeSetHashes())
			return
		case "/assets":
			assets, err := h.getLocalAssets()
			if err != nil {
				http.Error(w, "Couldn't get assets: "+err.Error(), 500)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, "%s", assets)
			return
		case "/head":
			w.Header()["Content-Type"] = []string{"text/plain"}
			fmt.Fprintf(w, "%s", h.getLedgerHeadString())
			return
		default:
			hash, err := ParseChangeSetHash(r.URL.Path)
			if err != nil {
				http.Error(w, "Invalid change set format", 400)
				return
			}

			cs, ok := h.ledger.GetChangeSet(hash)

			if !ok {
				http.Error(w, "unhandled sync GET path "+r.URL.Path, 404)
				return
			}

			bs, err := cs.Encode(false)

			if err != nil {
				http.Error(w, "Internal encoding error", 500)
				return
			}

			w.Header()["Content-Type"] = []string{"application/cbor"}
			w.Write(bs)
		}
	case "POST":
		switch r.URL.Path {
		case "/":
			defer r.Body.Close()
			body, err := io.ReadAll(r.Body)

			if err != nil {
				http.Error(w, "Empty body", 400)
				return
			}

			cs, err := DecodeChangeSet(body)
			if err != nil {
				http.Error(w, "Invalid change set: "+err.Error(), 400)
				return
			}

			if err := h.ledger.AppendChangeSet(cs, true); err != nil {
				http.Error(w, "Invalid change set: "+err.Error(), 400)
				return
			}

			if err := cs.Apply(h.resourceManager); err != nil {
				http.Error(w, "Invalid change set: "+err.Error(), 400)
				return
			}

			h.ledger.Write()
			fmt.Fprintf(w, "")
		default:
			http.Error(w, "unhandled sync POST path "+r.URL.Path, 404)
		}
	case "PUT":
		switch r.URL.Path {
		case "/assets":
			defer r.Body.Close()

			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Empty body", 400)
				return
			}

			id := GenerateAssetId(body)
			assetsDir := GetAssetsDir()

			path := assetsDir + "/" + id

			if _, err := os.Stat(path); err != nil {
				if errors.Is(err, os.ErrNotExist) {
					if err := os.WriteFile(path, body, 0644); err != nil {
						http.Error(w, "write error", 500)
						return
					}
				} else {
					http.Error(w, path+" error", 500)
					return
				}
			}

			fmt.Fprintf(w, "%s", id)
		}
	default:
		http.Error(w, "unsupported sync http method "+r.Method, 404)
	}
}

func (h *syncHandler) getLedgerHeadString() string {
	return StringifyChangeSetHash(h.ledger.Head)
}

func (h *syncHandler) getLedgerChangeSetHashes() string {
	return h.ledger.GetChangeSetHashes().Stringify()
}

func (h *syncHandler) getLocalAssets() (string, error) {
	assets := []string{}

	assetsDir := GetAssetsDir()
	files, err := os.ReadDir(assetsDir)

	if err != nil {
		return "", err
	}

	for _, f := range files {
		if strings.HasPrefix(f.Name(), "asset") {
			assets = append(assets, f.Name())
		}
	}

	res, err := json.Marshal(assets)
	if err != nil {
		return "", err
	}

	return string(res), nil
}
