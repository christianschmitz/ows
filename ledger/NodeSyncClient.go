package ledger

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
)

type NodeSyncClient struct {
	address string
}

func NewNodeSyncClient(address string) *NodeSyncClient {
	return &NodeSyncClient{address}
}

func (c *NodeSyncClient) url(relPath string) string {
	return "http://" + c.address + ":" + strconv.Itoa(SYNC_PORT) + "/" + relPath
}

func (c *NodeSyncClient) GetHead() (ChangeSetHash, error) {
	resp, err := http.Get(c.url("head"))

	if err != nil {
		return []byte{}, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}

	return ParseChangeSetHash(string(body))
}

func (c *NodeSyncClient) GetChangeSet(h ChangeSetHash) (*ChangeSet, error) {
	resp, err := http.Get(c.url(StringifyChangeSetHash(h)))

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return DecodeChangeSet(body)
}

func (c *NodeSyncClient) GetChangeSetHashes() (*ChangeSetHashes, error) {
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

	hashes := make([]ChangeSetHash, len(rawHashes))

	for i, rh := range rawHashes {
		h, err := ParseChangeSetHash(rh)

		if err != nil {
			return nil, err
		}

		hashes[i] = h
	}

	return &ChangeSetHashes{hashes}, nil
}

func (c *NodeSyncClient) PublishChangeSet(cs *ChangeSet) error {
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