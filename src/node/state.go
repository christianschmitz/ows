package main

import (
	"errors"
	"fmt"
	"os"
	"path"

	"ows/ledger"
	"ows/resources"
)

const (
	AppDirName           = "ows"
	AssetsDirName        = "assets"
	DefaultConfigDirName = "/etc"
	DefaultDataDirName   = "/var/lib"
	DefaultLogDirName    = "/var/log"
	KeyPairFileName      = "key"
	LedgerFileName       = "ledger"
	TestLogDirName       = "log"
)

type nodeState struct {
	testDir string

	cachedKeyPair *ledger.KeyPair
	cachedLedger  *ledger.Ledger

	resources *resources.Manager
}

func (s *nodeState) AddAsset(bs []byte) (ledger.AssetID, error) {
	return s.resources.Assets.Add(bs)
}

// Append the change set to the ledger, then write the ledger to disk, and
// finally sync the resources.
func (s *nodeState) AppendChangeSet(cs *ledger.ChangeSet) error {
	l := s.ledger()

	if err := l.Append(cs); err != nil {
		return err
	}

	if err := l.Write(s.ledgerPath()); err != nil {
		return err
	}

	if err := s.resources.Sync(l.Snapshot); err != nil {
		return err
	}

	// TODO: gossip to other nodes

	return nil
}

func (s *nodeState) Ledger() *ledger.Ledger {
	return s.ledger()
}

func (s *nodeState) ListAssets() []ledger.AssetID {
	return s.resources.Assets.ListAssets()
}

func (s *nodeState) Rollback(p int) error {
	l := s.ledger()

	l.Keep(p)

	return l.Write(s.ledgerPath())
}

func (s *nodeState) appConfigPath() string {
	return s.appPath(s.systemConfigPath())
}

func (s *nodeState) appDataPath() string {
	return s.appPath(s.systemDataPath())
}

func (s *nodeState) appLogPath() string {
	return s.appPath(s.systemLogPath())
}

func (s *nodeState) appPath(systemPath string) string {
	if s.testDir != "" {
		return systemPath
	} else {
		return path.Join(systemPath, AppDirName)
	}
}

func (s *nodeState) assetsPath() string {
	return path.Join(s.appDataPath(), AssetsDirName)
}

func (s *nodeState) keyPairPath() string {
	return path.Join(s.appConfigPath(), KeyPairFileName)
}

func (s *nodeState) ledgerPath() string {
	return path.Join(s.appDataPath(), LedgerFileName)
}

func (s *nodeState) systemConfigPath() string {
	if s.testDir != "" {
		kp, exists := ledger.EnvKeyPair()
		if !exists {
			panic(fmt.Sprintf("%s not set (must be set when --test-dir is set)", ledger.PrivateKeyEnvName))
		}

		nodeID := kp.Public.NodeID()

		return path.Join(s.testDir, string(nodeID))
	} else {
		return DefaultConfigDirName
	}
}

func (s *nodeState) systemDataPath() string {
	if s.testDir != "" {
		return s.systemConfigPath()
	} else {
		return DefaultDataDirName
	}
}

func (s *nodeState) systemLogPath() string {
	if s.testDir != "" {
		p := s.systemConfigPath()

		return path.Join(p, TestLogDirName)
	} else {
		return DefaultLogDirName
	}
}

func (s *nodeState) keyPair() *ledger.KeyPair {
	if s.cachedKeyPair != nil {
		return s.cachedKeyPair
	}

	p := s.keyPairPath()
	kp, existsInEnv := ledger.EnvKeyPair()

	if s.testDir != "" {
		if !existsInEnv {
			panic(fmt.Sprintf("%s not set (must be set when --test-dir is set)", ledger.PrivateKeyEnvName))
		}
	} else if !existsInEnv {
		var err error
		kp, err = ledger.ReadKeyPair(p)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				panic(fmt.Sprintf("node key not found at %s"))
			} else {
				panic(err)
			}
		}
	} else {
		// write key to disk if it isn't available
		if _, err := os.Stat(p); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				if err := kp.Write(p); err != nil {
					panic(fmt.Sprintf("unable to write key to %s (%v)", p, err))
				}
			} else {
				panic(fmt.Sprintf("unable to read existing key at %s (%v)", p, err))
			}
		}
	}

	s.cachedKeyPair = kp

	return kp
}

func (s *nodeState) ledger() *ledger.Ledger {
	if s.cachedLedger != nil {
		return s.cachedLedger
	}

	l, existsInEnv := ledger.EnvLedger()

	if !existsInEnv {
		lDisk, err := ledger.ReadLedger(s.ledgerPath())
		if err != nil {
			panic(fmt.Sprintf("unable to read ledger (%v)", err))
		} else {
			l = lDisk
		}
	} else {
		p := s.ledgerPath()

		lDisk, err := ledger.ReadLedger(p)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				if err := l.Write(p); err != nil {
					panic(err)
				}
			} else {
				panic(err)
			}
		} else if lDisk.ProjectID() != l.ProjectID() {
			panic(fmt.Sprintf("project id of ledger at %s (%s) doesn't correspond to env (%s)", p, lDisk.ProjectID(), l.ProjectID()))
		} else {
			l = lDisk
		}
	}

	s.cachedLedger = l

	return l
}
