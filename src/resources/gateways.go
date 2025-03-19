package resources

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"slices"
	"time"

	"ows/ledger"
)

type Endpoint struct {
	Config ledger.GatewayEndpointConfig
}

type Gateway struct {
	Port    ledger.Port
	Handler *GatewayHandler
	Server  *http.Server
}

type GatewayHandler struct {
	Functions *FunctionManager
	// first key is method: "GET", "POST", "DELETE", "PUT", "PATCH"
	// second key is relative path, including initial slash (eg. "/assets")
	Endpoints map[string]map[string]*Endpoint
}

type GatewaysManager struct {
	Functions *FunctionManager
	Gateways  map[ledger.GatewayID]*Gateway
}

func (g *Gateway) shutdown() error {
	fmt.Printf("Shutting down gateway at port %d\n", g.Port)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	return g.Server.Shutdown(ctx)
}

func (h *GatewayHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if endpoints, ok := h.Endpoints[r.Method]; ok {
		if endpoint, ok := endpoints[r.URL.Path]; ok {
			// now run the task
			resp, err := h.Functions.Run(endpoint.Config.FunctionID, "hello world")
			if err != nil {
				http.Error(w, fmt.Sprintf("failed to run task (%v)", err), 500)
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

func newGatewaysManager(functions *FunctionManager) *GatewaysManager {
	return &GatewaysManager{
		Functions: functions,
		Gateways:  map[ledger.GatewayID]*Gateway{},
	}
}

func (m *GatewaysManager) Sync(gateways map[ledger.GatewayID]ledger.GatewayConfig) error {
	for id, conf := range gateways {
		if _, ok := m.Gateways[id]; ok {
			if err := m.update(id, conf); err != nil {
				return fmt.Errorf("failed to update gateway %s (%v)", id, err)
			}
		} else {
			if err := m.add(id, conf); err != nil {
				return fmt.Errorf("failed to add gateway %s (%v)", id, err)
			}
		}
	}

	for id, _ := range m.Gateways {
		if _, ok := gateways[id]; !ok {
			if err := m.remove(id); err != nil {
				return fmt.Errorf("failed to remove gateway %s (%v)", id, err)
			}
		}
	}

	return nil
}

func (m *GatewaysManager) add(id ledger.GatewayID, config ledger.GatewayConfig) error {
	if _, ok := m.Gateways[id]; ok {
		return fmt.Errorf("gateway %s already exists", id)
	}

	h := &GatewayHandler{
		Functions: m.Functions,
		Endpoints: map[string]map[string]*Endpoint{},
	}

	s := &http.Server{
		Addr:           fmt.Sprintf(":%d", config.Port),
		Handler:        h,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	m.Gateways[id] = &Gateway{
		Port:    config.Port,
		Handler: h,
		Server:  s,
	}

	// TODO: flexible TLS, using DomainManager + LetsEncrypt
	go s.ListenAndServe()

	log.Printf("added gateway %s on port %d\n", id, config.Port)

	for _, ep := range config.Endpoints {
		if err := m.addEndpoint(id, ep); err != nil {
			return err
		}
	}

	return nil
}

func (m *GatewaysManager) remove(id ledger.GatewayID) error {
	gateway, ok := m.Gateways[id]
	if !ok {
		return fmt.Errorf("gateway %s not found", id)
	}

	err := gateway.shutdown()
	if err != nil {
		return err
	}

	delete(m.Gateways, id)

	log.Printf("removed gateway %s on port %d\n", id, gateway.Port)

	return nil
}

func (m *GatewaysManager) update(id ledger.GatewayID, config ledger.GatewayConfig) error {
	prev, ok := m.Gateways[id]
	if !ok {
		return fmt.Errorf("gateway %s not found", id)
	}

	// make sure the endpoints correspond

	// TODO: support port changes

	for _, ep := range config.Endpoints {
		if methodEndpoints, ok := prev.Handler.Endpoints[ep.Method]; ok {
			if _, ok := methodEndpoints[ep.Path]; ok {
				return m.updateEndpoint(id, ep)
			} else {
				return m.addEndpoint(id, ep)
			}
		} else {
			return m.addEndpoint(id, ep)
		}
	}

	for method, methodEndpoints := range prev.Handler.Endpoints {
		methodEndpointConfigs := slices.DeleteFunc(config.Endpoints, func(ep ledger.GatewayEndpointConfig) bool {
			return ep.Method != method
		})

		if len(methodEndpointConfigs) == 0 {
			for path, _ := range methodEndpoints {
				return m.removeEndpoint(id, method, path)
			}
		}

		for path, _ := range methodEndpoints {
			pathEndpointConfigs := slices.DeleteFunc(methodEndpointConfigs, func(ep ledger.GatewayEndpointConfig) bool {
				return ep.Path != path
			})

			if len(pathEndpointConfigs) == 0 {
				return m.removeEndpoint(id, method, path)
			}
		}
	}

	return nil
}

func (m *GatewaysManager) addEndpoint(gatewayID ledger.GatewayID, config ledger.GatewayEndpointConfig) error {
	gateway, ok := m.Gateways[gatewayID]
	if !ok {
		return fmt.Errorf("invalid gateway id %s", gatewayID)
	}

	method := config.Method
	relPath := config.Path

	endpoints, ok := gateway.Handler.Endpoints[method]
	if !ok {
		endpoints = map[string]*Endpoint{}
		gateway.Handler.Endpoints[method] = endpoints
	}

	if _, ok := endpoints[relPath]; ok {
		return errors.New("endpoint already exists")
	}

	endpoints[relPath] = &Endpoint{
		Config: config,
	}

	log.Printf("added endpoint %s %s to gateway %s (port %d)\n", config.Method, config.Path, gatewayID, gateway.Port)

	return nil
}

func (m *GatewaysManager) removeEndpoint(gatewayID ledger.GatewayID, method string, path string) error {
	gateway, ok := m.Gateways[gatewayID]
	if !ok {
		return fmt.Errorf("invalid gateway id %s", gatewayID)
	}

	endpoints, ok := gateway.Handler.Endpoints[method]
	if !ok {
		return fmt.Errorf("no endpoints with method %s found", method)
	}

	if _, ok := endpoints[path]; !ok {
		return fmt.Errorf("no endpoint with method %s and path %s found", method, path)
	}

	delete(endpoints, path)

	return nil
}

func (m *GatewaysManager) updateEndpoint(gatewayID ledger.GatewayID, config ledger.GatewayEndpointConfig) error {
	gateway, ok := m.Gateways[gatewayID]
	if !ok {
		return fmt.Errorf("invalid gateway id %s", gatewayID)
	}

	endpoints, ok := gateway.Handler.Endpoints[config.Method]
	if !ok {
		return fmt.Errorf("no endpoints with method %s found", config.Method)
	}

	if _, ok := endpoints[config.Path]; !ok {
		return fmt.Errorf("no endpoint with method %s and path %s found", config.Method, config.Path)
	}

	endpoints[config.Path] = &Endpoint{
		Config: config,
	}

	return nil
}
