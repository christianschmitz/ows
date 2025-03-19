package network

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"ows/ledger"
)

// General API client
type APIClient struct {
	kp        *ledger.KeyPair
	callbacks Callbacks
}

// Node-specific API client
type NodeAPIClient struct {
	httpClient *http.Client
	address    string
	port       ledger.Port
}

// Creates a new general API client
func NewAPIClient(kp *ledger.KeyPair, callbacks Callbacks) *APIClient {
	return &APIClient{kp, callbacks}
}

// Returns a node-specific API client
func (c *APIClient) PickNode() *NodeAPIClient {
	m := c.callbacks.Ledger().Snapshot.Nodes

	for _, conf := range m {
		return NewNodeAPIClient(c.kp, conf.Address, conf.APIPort, c.callbacks)
	}

	panic("no nodes found")
}

// Syncs the local ledger, by performing the following steps:
//  1. Pick any node
//  2. Request the head of that node's ledger
//  3. If the head is the same exit
//  4. If the head isn't the same, fetch all tx Ids
//  5. Find the intersection (last common point)
//  6. Download everything after the intersection
func (c *APIClient) Sync() error {
	node := c.PickNode()

	head, err := node.Head()
	if err != nil {
		return err
	}

	if c.callbacks.Ledger().Head() == head {
		return nil
	}

	remoteChangeSetIDs, err := node.ChangeSetIDChain()
	if err != nil {
		return err
	}

	thisChangeSetIDs := c.callbacks.Ledger().IDChain()

	p, err := thisChangeSetIDs.Intersect(remoteChangeSetIDs)
	if err != nil {
		return err
	}

	// remove [p+1:] from local ledger
	if err := c.callbacks.Rollback(p); err != nil {
		return err
	}

	// download [p+1:] from remote ledger
	if p+1 < len(remoteChangeSetIDs.IDs) {
		for i := int(p) + 1; i < len(remoteChangeSetIDs.IDs); i++ {
			h := remoteChangeSetIDs.IDs[i]

			cs, err := node.ChangeSet(h)
			if err != nil {
				return err
			}

			if err := c.callbacks.AppendChangeSet(cs); err != nil {
				return err
			}
		}
	}

	return nil
}

func NewNodeAPIClient(kp *ledger.KeyPair, address string, port ledger.Port, callbacks Callbacks) *NodeAPIClient {
	cert, err := makeTLSCertificate(*kp)
	if err != nil {
		panic(err)
	}

	tlsConf := makeClientTLSConfig(cert, func(peer ledger.PublicKey) bool {
		l := callbacks.Ledger()

		if _, ok := l.Snapshot.Nodes[peer.NodeID()]; ok {
			return true
		} else {
			return false
		}
	})

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConf,
		},
	}

	return &NodeAPIClient{httpClient, address, port}
}

func (c *NodeAPIClient) Assets() ([]ledger.AssetID, error) {
	resp, err := handleResponse(c.httpClient.Get(c.url("assets")))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	assets := []ledger.AssetID{}

	err = json.Unmarshal(body, &assets)
	if err != nil {
		return nil, err
	}

	return assets, nil
}

func (c *NodeAPIClient) ChangeSet(id ledger.ChangeSetID) (*ledger.ChangeSet, error) {
	resp, err := handleResponse(c.httpClient.Get(c.url(string(id))))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// TODO: get version from another endpoint
	return ledger.DecodeChangeSet(body, ledger.LatestLedgerVersion)
}

func (c *NodeAPIClient) ChangeSetIDChain() (*ledger.ChangeSetIDChain, error) {
	resp, err := handleResponse(c.httpClient.Get(c.url("")))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	rawIDs := []string{}

	err = json.Unmarshal(body, &rawIDs)
	if err != nil {
		return nil, err
	}

	ids := make([]ledger.ChangeSetID, len(rawIDs))

	for i, id := range rawIDs {
		if err := ledger.ValidateID(id, ledger.ChangeSetIDPrefix); err != nil {
			return nil, err
		}

		ids[i] = ledger.ChangeSetID(id)
	}

	return &ledger.ChangeSetIDChain{ids}, nil
}

func (c *NodeAPIClient) Head() (ledger.ChangeSetID, error) {
	resp, err := handleResponse(c.httpClient.Get(c.url("head")))
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	id := string(body)

	if err := ledger.ValidateID(id, ledger.ChangeSetIDPrefix); err != nil {
		return "", err
	}

	return ledger.ChangeSetID(id), nil
}

// Not responsible for appending locally via c.callbacks.AppendChangeSet()
func (c *NodeAPIClient) AppendChangeSet(cs *ledger.ChangeSet) error {
	bs := cs.Encode()

	if _, err := handleResponse(c.httpClient.Post(c.url(""), "application/cbor", bytes.NewBuffer(bs))); err != nil {
		return err
	}

	return nil
}

func (c *NodeAPIClient) UploadAsset(bs []byte) (ledger.AssetID, error) {
	req, err := http.NewRequest("PUT", c.url("assets"), bytes.NewBuffer(bs))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := handleResponse(c.httpClient.Do(req))
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	id := string(body)

	if err := ledger.ValidateID(id, ledger.AssetIDPrefix); err != nil {
		return ledger.AssetID(id), err
	}

	return ledger.AssetID(id), nil
}

func (c *NodeAPIClient) url(relPath string) string {
	return fmt.Sprintf("https://%s:%d/%s", c.address, c.port, relPath)
}

func handleResponse(resp *http.Response, err error) (*http.Response, error) {
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("request failed with code %d (%s)", resp.StatusCode, resp.Status)
		}

		return nil, fmt.Errorf("request failed (%s)", string(body))
	}

	return resp, nil
}
