package ledger

import (
	"fmt"
	"io"
	"log"
	"net/http"
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
			hash := ParseChangeSetHash(r.URL.Path)

			cs, ok := h.ledger.GetChangeSet(hash)

			if (!ok) {
				http.Error(w, "unhandled sync GET path " + r.URL.Path, 404)
				return
			}

			w.Header()["Content-Type"] = []string{"application/cbor"}
			w.Write(cs.Encode())
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
				fmt.Println(err)
				http.Error(w, "Invalid change set", 400)
				return
			}

			h.ledger.AppendChangeSet(cs)
			h.ledger.Persist()
			fmt.Fprintf(w, "")
		default:
			http.Error(w, "unhandled sync POST path " + r.URL.Path, 404)
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