package main

import (
	"errors"
	"fmt"
	"os"
	"path"

	"ows/ledger"
	"ows/network"
	"ows/resources"
)

const (
	HomePathEnvName      = "HOME"
	XDGCachePathEnvName  = "XDG_CACHE_HOME"
	XDGConfigPathEnvName = "XDG_CONFIG_HOME"
	XDGDataPathEnvName   = "XDG_DATA_HOME"

	AppDirName           = "ows"
	AssetsDirName        = "assets"
	DefaultCacheDirName  = ".cache"
	DefaultConfigDirName = ".config"
	DefaultDataDirName   = ".local/share"
	KeyPairFileName      = "key"
	LedgerFileName       = "ledger"
	LogsDirName          = "logs"
	ProjectsDirName      = "projects"
)

// The clientState is responsible for resolving files
type clientState struct {
	// don't sync with nodes if `isOffline` is true
	isOffline bool

	projectName string
	testDir     string

	cachedKeyPair *ledger.KeyPair
	cachedLedger  *ledger.Ledger
}

func (s *clientState) AddAsset(bs []byte) (ledger.AssetID, error) {
	return resources.AddAsset(s.assetsPath(), bs)
}

func (s *clientState) AppendChangeSet(cs *ledger.ChangeSet) error {
	l := s.ledger()

	if err := l.Append(cs); err != nil {
		return err
	}

	if err := l.Write(s.ledgerPath()); err != nil {
		return err
	}

	return nil
}

func (s *clientState) Ledger() *ledger.Ledger {
	return s.ledger()
}

func (s *clientState) ListAssets() []ledger.AssetID {
	return resources.ListAssets(s.assetsPath())
}

func (s *clientState) Rollback(p int) error {
	l := s.ledger()

	l.Keep(p)

	return l.Write(s.ledgerPath())
}

func (s *clientState) appCachePath() string {
	return s.appPath(s.userCachePath())
}

func (s *clientState) appConfigPath() string {
	return s.appPath(s.userConfigPath())
}

func (s *clientState) appDataPath() string {
	return s.appPath(s.userDataPath())
}

// Only add AppDirName to `userPath` if --test-dir isn't set (if --test-dir
// is set this would be a redundant additional directory level).
func (s *clientState) appPath(userPath string) string {
	if s.testDir != "" {
		return userPath
	} else {
		return path.Join(userPath, AppDirName)
	}
}

func (s *clientState) assetsPath() string {
	return path.Join(s.appCachePath(), AssetsDirName)
}

func (s *clientState) currentProjectPath() string {
	return path.Join(s.projectsDataPath(), string(s.currentProjectID()))
}

func (s *clientState) keyPairPath() string {
	return path.Join(s.appConfigPath(), KeyPairFileName)
}

func (s *clientState) ledgerPath() string {
	return path.Join(s.currentProjectPath(), LedgerFileName)
}

func (s *clientState) logsPath() string {
	return path.Join(s.appCachePath(), LogsDirName)
}

func (s *clientState) projectsConfigPath() string {
	return path.Join(s.appConfigPath(), ProjectsDirName)
}

func (s *clientState) projectsDataPath() string {
	return path.Join(s.appDataPath(), ProjectsDirName)
}

func (s *clientState) appendActions(actions ...ledger.Action) error {
	cs := s.ledger().NewChangeSet(actions...)
	kp := s.keyPair()

	sig, err := kp.SignChangeSet(cs)
	if err != nil {
		return err
	}

	cs.Signatures = []ledger.Signature{sig}

	// Append locally
	if err := s.AppendChangeSet(cs); err != nil {
		return err
	}

	nc := s.newAPIClient().PickNode()

	// Append remotely
	return nc.AppendChangeSet(cs)
}

func (s *clientState) currentProjectID() ledger.ProjectID {
	if s.testDir != "" {
		l, exists := ledger.EnvLedger()
		if !exists {
			panic(fmt.Sprintf("%s not set (must be set when --test-dir is set)", ledger.InitialConfigEnvName))
		}

		return l.ProjectID()
	} else {
		// Map project name to projectID
		projectMappingPath := path.Join(s.projectsConfigPath(), s.projectName)
		bs, err := os.ReadFile(projectMappingPath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				panic(fmt.Sprintf("project %s not found at %s", s.projectName, projectMappingPath))
			} else {
				panic(err)
			}
		}

		return ledger.ProjectID(string(bs))
	}
}

func (s *clientState) userCachePath() string {
	if s.testDir != "" {
		return s.userConfigPath()
	}

	xdgCachePath, exists := os.LookupEnv(XDGCachePathEnvName)
	if exists {
		return xdgCachePath
	} else {
		homePath := envHomePath("unable to resolve cache path")
		return path.Join(homePath, DefaultCacheDirName)
	}
}

func (s *clientState) userConfigPath() string {
	if s.testDir != "" {
		kp := s.keyPair()

		userID := kp.Public.UserID()

		return path.Join(s.testDir, string(userID))
	}

	xdgConfigPath, exists := os.LookupEnv(XDGConfigPathEnvName)
	if exists {
		return xdgConfigPath
	} else {
		homePath := envHomePath("unable to resolve config path")
		return path.Join(homePath, DefaultConfigDirName)
	}
}

func (s *clientState) userDataPath() string {
	if s.testDir != "" {
		return s.userConfigPath()
	}

	xdgDataPath, exists := os.LookupEnv(XDGDataPathEnvName)
	if exists {
		return xdgDataPath
	} else {
		homePath := envHomePath("unable to resolve data path")
		return path.Join(homePath, DefaultDataDirName)
	}
}

func (s *clientState) keyPair() *ledger.KeyPair {
	if s.cachedKeyPair != nil {
		return s.cachedKeyPair
	}

	kp, existsInEnv := ledger.EnvKeyPair()

	if s.testDir != "" {
		if !existsInEnv {
			panic(fmt.Sprintf("%s not set (must be set when --test-dir is set)", ledger.PrivateKeyEnvName))
		}
	} else if !existsInEnv {
		p := s.keyPairPath()

		var err error
		kp, err = ledger.ReadKeyPair(p)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				panic(fmt.Sprintf("client key not found at %s (hint: use `ows key init` to create a new random key)", p))
			} else {
				panic(err)
			}
		}
	}

	s.cachedKeyPair = kp

	return kp
}

func (s *clientState) ledger() *ledger.Ledger {
	if s.cachedLedger != nil {
		return s.cachedLedger
	}

	l, existsInEnv := ledger.EnvLedger()

	if s.testDir != "" {
		if !existsInEnv {
			panic(fmt.Sprintf("%s not set (must be set when --test-dir is set)", ledger.InitialConfigEnvName))
		}
	} else if !existsInEnv {
		p := s.ledgerPath()

		var err error

		l, err = ledger.ReadLedger(p)
		if errors.Is(err, os.ErrNotExist) {
			panic(fmt.Sprintf("project ledger not found at %s", p))
		} else if err != nil {
			panic(err)
		}
	}

	s.cachedLedger = l

	if !s.isOffline {
		c := s.newAPIClient()

		if err := c.Sync(); err != nil {
			panic(err)
		}
	}

	return l
}

func (s *clientState) newAPIClient() *network.APIClient {
	return network.NewAPIClient(s.keyPair(), s)
}

func envHomePath(failMessage string) string {
	homePath, exists := os.LookupEnv(HomePathEnvName)
	if !exists {
		panic(fmt.Sprintf("%s env variable not set, %s", HomePathEnvName, failMessage))
	}

	return homePath
}

func saveKeyPair(kp *ledger.KeyPair) error {
	p := state.keyPairPath()

	// Make sure key doesn't already exist
	if _, err := os.Stat(p); err == nil {
		return fmt.Errorf("key already exists at %s", p)
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("an error occured while reading existing key at %s (%v)", p, err)
	}

	if err := kp.Write(p); err != nil {
		return fmt.Errorf("failure while writing key to %s (%v)", p, err)
	}

	return nil
}
