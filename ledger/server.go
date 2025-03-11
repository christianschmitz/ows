package ledger

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
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
			w.Header()["Content-Type"] = []string{"application/json"}
			fmt.Fprintf(w, "%s", h.getLedgerChangeSetHashes())
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

			if (!ok) {
				http.Error(w, "unhandled sync GET path " + r.URL.Path, 404)
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
				http.Error(w, "Invalid change set", 400)
				return
			}

			if err := h.ledger.AppendChangeSet(cs); err != nil {
				http.Error(w, "Invalid change set", 400)
				return
			}
			h.ledger.Write()
			fmt.Fprintf(w, "")
		default:
			http.Error(w, "unhandled sync POST path " + r.URL.Path, 404)
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

			assetsDir := HomeDir + "/assets"

			if err := os.MkdirAll(assetsDir, os.ModePerm); err != nil {
				http.Error(w, "Unable to make assets dir", 500)
				return
			}

			path := assetsDir + "/" + StringifyAssetId(id)

			if _, err := os.Stat(path); err != nil {
				if errors.Is(err, os.ErrNotExist) {
					if err := os.WriteFile(path, body, 0644); err != nil {
						http.Error(w, "write error", 500)
						return
					}
				} else {
					http.Error(w, path + " error", 500)
					return
				}
			}

			fmt.Fprintf(w, "%s", StringifyAssetId(id))
		}
	default:
		http.Error(w, "unsupported sync http method " + r.Method, 404)
	}
}

func (h *syncHandler) getLedgerHeadString() string {
	return StringifyChangeSetHash(h.ledger.Head)
}

func (h *syncHandler) getLedgerChangeSetHashes() string {
	return h.ledger.GetChangeSetHashes().Stringify()
}