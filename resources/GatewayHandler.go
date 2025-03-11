package resources

import (
	"fmt"
	"net/http"
	"ows/ledger"
)

type GatewayHandler struct {
	Tasks *TasksManager
	// first key is method: "GET", "POST", "DELETE", "PUT", "PATCH"
	// second key is relative path, including initial slash (eg. "/assets")
	Endpoints map[string]map[string]EndpointConfig
}

// TODO: timeout details
type EndpointConfig struct {
	Task ledger.ResourceId // task to run
}

func (h *GatewayHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if endpoints, ok := h.Endpoints[r.Method]; ok {
		if endpoint, ok := endpoints[r.URL.Path]; ok {
			// now run the task
			str, err := h.Tasks.Run(endpoint.Task)
			if err != nil {
				http.Error(w, "failed to run task", 500)
				return
			}

			fmt.Fprintf(w, str)
		} else {
			http.Error(w, "invalid path", 404)	
		}
	} else {
		http.Error(w, "unsupported method", 404)
	}
}