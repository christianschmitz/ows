package resources

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type GatewayHandler struct {
	Tasks *TasksManager
	// first key is method: "GET", "POST", "DELETE", "PUT", "PATCH"
	// second key is relative path, including initial slash (eg. "/assets")
	Endpoints map[string]map[string]EndpointConfig
}

// TODO: timeout details
type EndpointConfig struct {
	TaskId string // task to run
}

func (h *GatewayHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if endpoints, ok := h.Endpoints[r.Method]; ok {
		if endpoint, ok := endpoints[r.URL.Path]; ok {
			// now run the task
			resp, err := h.Tasks.Run(endpoint.TaskId, "hello world")
			if err != nil {
				fmt.Println(err)
				http.Error(w, "failed to run task", 500)
				return
			}

			str, err := json.Marshal(resp)
			if err != nil {
				http.Error(w, "bad response", 500)
				return
			}

			fmt.Fprintf(w, string(str))
		} else {
			http.Error(w, "invalid path", 404)	
		}
	} else {
		http.Error(w, "unsupported method", 404)
	}
}