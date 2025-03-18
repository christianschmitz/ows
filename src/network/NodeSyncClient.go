package network

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"ows/ledger"
)

type NodeSyncClient struct {
	address string
	port    ledger.Port
}

func NewNodeSyncClient(address string, port ledger.Port) *NodeSyncClient {
	return &NodeSyncClient{address, port}
}

func (c *NodeSyncClient) url(relPath string) string {
	return fmt.Sprintf("http://%s:%d/%s", c.address, c.port, relPath)
}

func (c *NodeSyncClient) Head() (ledger.ChangeSetID, error) {
	resp, err := http.Get(c.url("head"))
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

func (c *NodeSyncClient) ChangeSet(id ledger.ChangeSetID) (*ledger.ChangeSet, error) {
	resp, err := http.Get(c.url(string(id)))
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

func (c *NodeSyncClient) ChangeSetIDChain() (*ledger.ChangeSetIDChain, error) {
	resp, err := http.Get(c.url(""))
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

func (c *NodeSyncClient) PublishChangeSet(cs *ledger.ChangeSet) error {
	bs := cs.Encode()

	resp, err := http.Post(c.url(""), "application/cbor", bytes.NewBuffer(bs))
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return errors.New("request failed: " + resp.Status)
		}

		return errors.New(string(body))
	}

	return nil
}

func (c *NodeSyncClient) UploadFile(bs []byte) (string, error) {
	req, err := http.NewRequest("PUT", c.url("assets"), bytes.NewBuffer(bs))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", errors.New("Upload error " + resp.Status)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func (c *NodeSyncClient) GetAssets() ([]string, error) {
	resp, err := http.Get(c.url("assets"))
	if err != nil {
		return []string{}, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, errors.New(string(body))
	}

	assets := []string{}

	err = json.Unmarshal(body, &assets)
	if err != nil {
		return nil, err
	}

	return assets, nil
}
