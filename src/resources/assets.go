package resources

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"ows/ledger"
)

type AssetManager struct {
	AssetsDir string
}

func newAssetManager(assetsDir string) *AssetManager {
	return &AssetManager{
		AssetsDir: assetsDir,
	}
}

func (m *AssetManager) Add(bs []byte) (ledger.AssetID, error) {
	id := ledger.GenerateAssetID(bs)

	p := path.Join(m.AssetsDir, string(id))

	if _, err := os.Stat(p); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if err := os.WriteFile(p, bs, 0644); err != nil {
				return id, fmt.Errorf("write error (%v)", err)
			}
		} else {
			return id, fmt.Errorf("failed to read preexisting asset (%v)", err)
		}
	}

	return id, nil

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

// List all locally stored assets
func (m *AssetManager) ListAssets() []ledger.AssetID {
	assets := make([]ledger.AssetID, 0)

	files, err := os.ReadDir(m.AssetsDir)
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
