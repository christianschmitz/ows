package sync

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

func (c *NodeSyncClient) GetHead() (ledger.ChangeSetHash, error) {
	resp, err := http.Get(c.url("head"))

	if err != nil {
		return []byte{}, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}

	return ledger.ParseChangeSetHash(string(body))
}

func (c *NodeSyncClient) GetChangeSet(h ledger.ChangeSetHash) (*ledger.ChangeSet, error) {
	resp, err := http.Get(c.url(ledger.StringifyChangeSetHash(h)))
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

func (c *NodeSyncClient) GetChangeSetHashes() (*ledger.ChangeSetHashes, error) {
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

	hashes := make([]ledger.ChangeSetHash, len(rawHashes))

	for i, rh := range rawHashes {
		if (i == 0) {
			h, err := ledger.ParseProjectHash(rh)
			if err != nil {
				return nil, err
			}

			hashes[i] = h
		} else {
			h, err := ledger.ParseChangeSetHash(rh)
			if err != nil {
				return nil, err
			}

			hashes[i] = h
		}
	}

	return &ledger.ChangeSetHashes{hashes}, nil
}

func (c *NodeSyncClient) PublishChangeSet(cs *ledger.ChangeSet) error {
	bs, err := cs.Encode(false)
	if err != nil {
		return err
	}

	resp, err := http.Post(c.url(""), "application/cbor", bytes.NewBuffer(bs))
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return errors.New("request failed: " + resp.Status)		
	}

	return nil
}