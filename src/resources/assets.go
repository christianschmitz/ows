package resources

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"ows/ledger"
	"ows/network"
)

type AssetManager struct {
	Nodes     *NodeManager
	AssetsDir string
}

func newAssetManager(assetsDir string, nodes *NodeManager) *AssetManager {
	return &AssetManager{
		Nodes:     nodes,
		AssetsDir: assetsDir,
	}
}

func (m *AssetManager) Add(bs []byte) (ledger.AssetID, error) {
	return AddAsset(m.AssetsDir, bs)
}

func AddAsset(assetsDir string, bs []byte) (ledger.AssetID, error) {
	id := ledger.GenerateAssetID(bs)

	p := path.Join(assetsDir, string(id))

	log.Printf("writing asset %s to %s\n", id, p)

	if _, err := os.Stat(p); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if err := ledger.OverwriteSafe(p, bs); err != nil {
				return id, fmt.Errorf("write error (%v)", err)
			}
		} else {
			return id, fmt.Errorf("failed to read preexisting asset (%v)", err)
		}
	}

	return id, nil
}

// Get the asset locally.
// Returns an os.ErrNotExist error if not found
func (m *AssetManager) Get(id ledger.AssetID) ([]byte, error) {
	return GetAsset(m.AssetsDir, id)
}

func GetAsset(assetsDir string, id ledger.AssetID) ([]byte, error) {
	p := path.Join(assetsDir, string(id))

	bs, err := os.ReadFile(p)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Printf("asset %s not found locally at %p\n", id, p)
		}
		return nil, err
	}

	return bs, nil
}

func (m *AssetManager) AssetExists(id ledger.AssetID) bool {
	p := path.Join(m.AssetsDir, string(id))

	_, err := os.Stat(p)

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false
		} else {
			log.Println(err)
			return false
		}
	} else {
		// TODO: should we check if file is corrupt or not?
		return true
	}
}

// First looks for asset locally.
// Then tries to download from peers.
func (m *AssetManager) AssertExists(id ledger.AssetID) error {
	log.Println(m.AssetsDir)
	if m.AssetExists(id) {
		return nil
	}

	// download from the node that is nearest to the asset first
	otherNodeIDs := m.Nodes.OtherNodeIDs()
	network.SortNodesByDistanceToTarget(otherNodeIDs, string(id))

	var lastError error

	for _, otherNodeID := range otherNodeIDs {
		c, err := m.Nodes.NewClient(otherNodeID)
		if err != nil {
			lastError = err
			continue
		}

		bs, err := c.Asset(id)
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				lastError = err
			}

			continue
		}

		if _, err := m.Add(bs); err != nil {
			return err
		} else {
			return nil
		}
	}

	if lastError != nil {
		return fmt.Errorf("failed to connect to a node to download asset %s (%w)", id, lastError)
	}

	return fmt.Errorf("asset %s not found on peers (%w)", id, os.ErrNotExist)
}

// List all locally stored assets
func (m *AssetManager) ListAssets() []ledger.AssetID {
	return ListAssets(m.AssetsDir)
}

func ListAssets(assetsDir string) []ledger.AssetID {
	assets := make([]ledger.AssetID, 0)

	files, err := os.ReadDir(assetsDir)
	if err != nil {
		return []ledger.AssetID{}
	}

	for _, f := range files {
		name := f.Name()

		if strings.HasPrefix(name, ledger.AssetIDPrefix) {
			if err := ledger.ValidateID(name, ledger.AssetIDPrefix); err == nil { // NOT error
				assets = append(assets, ledger.AssetID(name))
			}
		}
	}

	return assets
}
