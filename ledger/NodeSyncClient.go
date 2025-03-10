package ledger

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
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

func (c *NodeSyncClient) GetHead() ChangeSetHash {
	resp, err := http.Get(c.url("head"))

	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)	
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

func (c *NodeSyncClient) GetChangeSetHashes() *ChangeSetHashes {
	resp, err := http.Get(c.url(""))

	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)	
	}

	rawHashes := []string{}

	err = json.Unmarshal(body, &rawHashes)

	if err != nil {
		log.Fatal(err)
	}

	hashes := make([]ChangeSetHash, len(rawHashes))

	for i, h := range rawHashes {
		hashes[i] = ParseChangeSetHash(h)
	}

	return &ChangeSetHashes{hashes}
}

func (c *NodeSyncClient) PublishChangeSet(cs *ChangeSet) {
	cs.Encode()

	resp, err := http.Post(c.url(""), "application/cbor", bytes.NewBuffer(cs.Encode()))

	if err != nil {
		log.Fatal(err)
	}

	if resp.StatusCode != 200 {
		log.Fatal("request failed: " + resp.Status)
	}
}