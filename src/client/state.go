package main

import (
	"errors"
	"fmt"
	"os"
	"path"

	"ows/ledger"
)

const (
	HomePathEnvName = "HOME"
	XDGCachePathEnvName = "XDG_CACHE_HOME"
	XDGConfigPathEnvName = "XDG_CONFIG_HOME"
	XDGDataPathEnvName = "XDG_DATA_HOME"

	AppDirName = "ows"
	DefaultCacheDirName = ".cache"
	DefaultConfigDirName = ".config"
	DefaultDataDirName = ".local/share"
	AssetsDirName = "assets"
	KeyPairFileName = "key"
	LedgerFileName = "ledger"
	LogsDirName = "logs"
	ProjectsDirName = "projects"
)

// The clientState is responsible for resolving files
type clientState struct {
	projectName string
	testDir string

	cachedKeyPair *ledger.KeyPair
	cachedLedger  *ledger.Ledger
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

func (s *clientState) keyPairPath() string {
	return path.Join(s.appConfigPath(), KeyPairFileName)
}

func (s *clientState) ledgerPath() string {
	return path.Join(s.projectPath(), LedgerFileName)
}

func (s *clientState) logsPath() string {
	return path.Join(s.appCachePath(), LogsDirName)
}

func (s *clientState) projectPath() string {
	return path.Join(s.appDataPath(), ProjectsDirName, string(s.projectID()))
}

func (s *clientState) projectID() ledger.ProjectID {
	if s.testDir != "" {
		l, exists := envLedger()
		if !exists {
			panic(fmt.Sprintf("%s not set (must be set when --test-dir is set)", ledger.InitialConfigEnvName))
		}

		return l.ProjectID()
	} else {
		// Map project name to projectID
		projectMappingPath := path.Join(s.userConfigPath(), ProjectsDirName, s.projectName)
		bs, err := os.ReadFile(projectMappingPath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				panic(fmt.Sprintf("%s doesn't exist", projectMappingPath))
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

	kp, existsInEnv := envKeyPair()

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

	l, existsInEnv := envLedger()

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
		} else {
			panic(err)
		}
	}

	s.cachedLedger = l

	return l
}

func envHomePath(failMessage string) string {
	homePath, exists := os.LookupEnv(HomePathEnvName)
	if !exists {
		panic(fmt.Sprintf("%s env variable not set, %s", HomePathEnvName, failMessage))
	}

	return homePath
}

func envKeyPair() (*ledger.KeyPair, bool) {
	privateKeyStr, exists := os.LookupEnv(ledger.PrivateKeyEnvName)
	if !exists {
		return nil, false
		
	}

	k, err := ledger.ParsePrivateKey(privateKeyStr)
	if err != nil {
		panic(fmt.Sprintf("invalid %s (%v)", ledger.PrivateKeyEnvName, err))
	}

	return k.KeyPair(), true
}

func envLedger() (*ledger.Ledger, bool) {
	initialConfigStr, exists := os.LookupEnv(ledger.InitialConfigEnvName)
	if !exists {
		return nil, false
	}

	l, err := ledger.ParseInitialConfig(initialConfigStr)
	if err != nil {
		panic(fmt.Sprintf("invalid %s (%v)", ledger.InitialConfigEnvName, err))
	}

	return l, true
}