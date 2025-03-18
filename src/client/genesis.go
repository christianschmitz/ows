package main

import (
//"bytes"
//"errors"
//"os"
//"strings"
)

//var GENESIS_PARENT_HASH = []byte{}
//
//const GENESIS_ENV_VAR_NAME = "OWS_GENESIS"

//func NewGenesisChangeSet(actions ...Action) *ChangeSet {
//	return &ChangeSet{GENESIS_PARENT_HASH, actions, []Signature{}}
//}

//func DecodeGenesisChangeSet(bs []byte) (*ledger.ChangeSet, error) {
//	cs, err := DecodeChangeSet(bs)
//	if err != nil {
//		return nil, err
//	}
//
//	if !bytes.Equal(GENESIS_PARENT_HASH, cs.Parent) {
//		return nil, errors.New("genesis parent id must be empty bytestring")
//	}
//
//	return cs, nil
//}

//func LookupGenesisChangeSet() (*ChangeSet, error) {
//	str, exists := os.LookupEnv(GENESIS_ENV_VAR_NAME)
//	if !exists {
//		defaultProfileChangeSet, ok := useDefaultProfileGenesis()
//		if ok {
//			return defaultProfileChangeSet, nil
//		}
//
//		singletonChangeSet, ok := useSingletonProjectGenesis()
//		if ok {
//			return singletonChangeSet, nil
//		}
//
//		return nil, errors.New(GENESIS_ENV_VAR_NAME + " is not set")
//	}
//
//	bs, err := DecodeBytes(str)
//	if err != nil {
//		return nil, err
//	}
//
//	g, err := DecodeGenesisChangeSet(bs)
//	if err != nil {
//		return nil, err
//	}
//
//	return g, nil
//}

//func useSingletonProjectGenesis() (*ledger.ChangeSet, bool) {
//	dir := HomeDir + "/"
//
//	files, err := os.ReadDir(dir)
//	if err != nil {
//		return nil, false
//	}
//
//	filteredPaths := make([]string, 0)
//
//	for _, file := range files {
//		if file.IsDir() && strings.HasPrefix(file.Name(), PROJECT_HASH_PREFIX) {
//			filteredPaths = append(filteredPaths, dir+file.Name())
//		}
//	}
//
//	if len(filteredPaths) != 1 {
//		return nil, false
//	}
//
//	projectPath := filteredPaths[0]
//
//	l, ok := readLedger(makeLedgerPath(projectPath))
//	if !ok {
//		return nil, false
//	}
//
//	return &(l.Changes[0]), true
//}

//func useDefaultProfileGenesis() (*ChangeSet, bool) {
//	path := HomeDir + "/profiles/default"
//
//	if _, err := os.Stat(path); err == nil {
//		bytes, err := os.ReadFile(path)
//		if err != nil {
//			return nil, false
//		}
//
//		str := string(bytes)
//
//		bs, err := DecodeBytes(str)
//		if err != nil {
//			return nil, false
//		}
//
//		cs, err := DecodeGenesisChangeSet(bs)
//		if err != nil {
//			return nil, false
//		}
//
//		return cs, true
//	} else {
//		return nil, false
//	}
//}
