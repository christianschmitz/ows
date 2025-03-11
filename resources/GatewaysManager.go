package resources

import (
	"errors"
	"net/http"
	"strconv"
	"time"
	"ows/ledger"
)

type Gateway struct {
	Port int
	Handler *GatewayHandler
	Server *http.Server
}

type GatewaysManager struct {
	Tasks *TasksManager
	Gateways map[string]*Gateway
}

func NewGatewaysManager(tm *TasksManager) *GatewaysManager {
	return &GatewaysManager{
		tm,
		map[string]*Gateway{},
	}
}

func (m *GatewaysManager) Add(id ledger.ResourceId, port int) error {
	sId := ledger.StringifyResourceId(id)

	h := &GatewayHandler{
		m.Tasks,
		map[string]map[string]EndpointConfig{},
	}

	s := &http.Server{
		Addr: ":" + strconv.Itoa(port),
		Handler: h,
		ReadTimeout: 10 * time.Second,
		WriteTimeout: 10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	if _, ok := m.Gateways[sId]; ok {
		return errors.New("gateway already exists")
	}

	m.Gateways[sId] = &Gateway{
		port,
		h,
		s,
	}

	go s.ListenAndServe()

	return nil
}

func (m *GatewaysManager) AddEndpoint(gatewayId ledger.ResourceId, method string, relPath string, task ledger.ResourceId) error {
	g, ok := m.Gateways[ledger.StringifyResourceId(gatewayId)]
	if !ok {
		return errors.New("invalid gateway id")
	}

	endpoints, ok := g.Handler.Endpoints[method]
	if !ok {
		endpoints = map[string]EndpointConfig{}
		g.Handler.Endpoints[method] = endpoints
	}

	if _, ok := endpoints[relPath]; ok {
		return errors.New("endpoint already exists")
	}

	endpoints[relPath] = EndpointConfig{
		task,
	}

	return nil
}