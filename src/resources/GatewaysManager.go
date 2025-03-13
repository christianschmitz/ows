package resources

import (
	"errors"
	"net/http"
	"strconv"
	"time"	
)

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

func (m *GatewaysManager) Add(id string, port int) error {
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

	if _, ok := m.Gateways[id]; ok {
		return errors.New("gateway already exists")
	}

	m.Gateways[id] = &Gateway{
		port,
		h,
		s,
	}

	go s.ListenAndServe()

	return nil
}

func (m *GatewaysManager) Remove(id string) error {
	g, ok := m.Gateways[id]
	if !ok {
		return errors.New("gateway " + id + " doesn't exist")
	}

	err := g.shutdown()
	if err != nil {
		return err
	}

	delete(m.Gateways, id)

	return nil
}

func (m *GatewaysManager) AddEndpoint(gatewayId string, method string, relPath string, taskId string) error {
	g, ok := m.Gateways[gatewayId]
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
		taskId,
	}

	return nil
}