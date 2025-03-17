package network

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"ows/ledger"
)

type NodeSyncClient struct {
	address string
}

func NewNodeSyncClient(address string) *NodeSyncClient {
	return &NodeSyncClient{address}
}

func (c *NodeSyncClient) url(relPath string) string {
	return "http://" + c.address + ":" + strconv.Itoa(ledger.SYNC_PORT) + "/" + relPath
}

func (c *NodeSyncClient) GetHead() (ledger.ChangeSetID, error) {
	resp, err := http.Get(c.url("head"))

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return ledger.ParseChangeSetID(string(body))
}

func (c *NodeSyncClient) GetChangeSet(h ledger.ChangeSetID) (*ledger.ChangeSet, error) {
	resp, err := http.Get(c.url(ledger.StringifyChangeSetID(h)))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return ledger.DecodeChangeSet(body)
}

func (c *NodeSyncClient) GetChangeSetIDs() (*ledger.ChangeSetIDs, error) {
	resp, err := http.Get(c.url(""))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	rawHashes := []string{}

	err = json.Unmarshal(body, &rawHashes)

	if err != nil {
		return nil, err
	}

	hashes := make([]ledger.ChangeSetID, len(rawHashes))

	for i, rh := range rawHashes {
		if i == 0 {
			h, err := ledger.ParseProjectHash(rh)
			if err != nil {
				return nil, err
			}

			hashes[i] = h
		} else {
			h, err := ledger.ParseChangeSetID(rh)
			if err != nil {
				return nil, err
			}

			hashes[i] = h
		}
	}

	return &ledger.ChangeSetIDChain{hashes}, nil
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
