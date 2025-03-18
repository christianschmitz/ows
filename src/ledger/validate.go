package ledger

import (
	"fmt"
)

// Instead of passing a plain list of change sets into validateAllChangeSets(),
// we pass a generator, so that validation can be done whilst decoding.
//
// Decoding and validation are intertwined due to potential ledger version
// changes, and ledger version changes can result in ledger ChangeSet codec
// changes.
type changeSetGenerator = func(yield func(cs *ChangeSet, err error) bool)

// (Re)validates the ledger from the beginning, and recreates the snapshot
func (l *Ledger) Validate() error {
	snapshot := newSnapshot(l.InitialVersion)

	genChanges := func(yield func(cs *ChangeSet, err error) bool) {
		i := 1

		for {
			if i < len(l.Changes) {
				cs := l.Changes[i]

				i++

				if !yield(&cs, nil) {
					return
				}
			} else {
				return
			}
		}
	}

	_, err := validateAllChangeSets(&(l.Changes[0]), genChanges, snapshot)
	if err != nil {
		return fmt.Errorf("failed to validate ledger (%v)", err)
	}

	l.Snapshot = snapshot

	return nil
}

// Uses a generator function for the changes, so this function can be used in
// different situations.
func validateAllChangeSets(
	initialConfig *ChangeSet,
	genChanges changeSetGenerator,
	snapshot *Snapshot,

) ([]ChangeSet, error) {
	if err := validateFirstChangeSet(initialConfig, snapshot); err != nil {
		return nil, err
	}

	changes := []ChangeSet{*initialConfig}

	for cs := range genChanges {
		if err := validateChangeSet(cs, snapshot); err != nil {
			return nil, fmt.Errorf("failed to validate change set (%v)", err)
		}

		changes = append(changes, *cs)
	}

	return changes, nil
}

// Validates the initial project configuration.
func validateFirstChangeSet(cs *ChangeSet, snapshot *Snapshot) error {
	if cs.Prev != "" {
		return fmt.Errorf("invalid Prev ChangeSetID for first change set, expected empty string, got %s", cs.Prev)
	}

	rootUsers, err := cs.validateSignatures()
	if err != nil {
		return fmt.Errorf("invalid root signature (%v)", err)
	}

	snapshot.addRootUsers(rootUsers...)

	if err := cs.apply(snapshot); err != nil {
		return fmt.Errorf("failed to validate first change set (%v)", err)
	}

	snapshot.Head = cs.ID()

	return nil
}

func validateChangeSet(cs *ChangeSet, snapshot *Snapshot) error {
	if cs.Prev != snapshot.Head {
		return fmt.Errorf("invalid Prev ChangeSetID, expected %s, got %s", snapshot.Head, cs.Prev)
	}

	signers, err := cs.validateSignatures()
	if err != nil {
		return err
	}

	// check that all the actions can actually be taken by the signers
	policies, err := snapshot.UserPolicies(signers)
	if err != nil {
		return err
	}

	for _, a := range cs.Actions {
		if !actionAllowed(a, policies...) {
			return fmt.Errorf("merged policy of all signers doesn't allow %s:%s", a.Category(), a.Name())
		}
	}

	if err := cs.apply(snapshot); err != nil {
		return err
	}

	snapshot.Head = cs.ID()

	return nil
}

func validateResourceID(id ResourceID, expectedPrefix string) error {
	prefix, bs, err := DecodeBech32(string(id))
	if err != nil {
		return fmt.Errorf("invalid resource id format %s (%v)", id, err)
	}

	if len(bs) != shortDigestSize {
		return fmt.Errorf("invalid number of bytes in %s", id)
	}

	if prefix != expectedPrefix {
		return fmt.Errorf("invalid resource id %s, expected prefix %s, got %s", id, expectedPrefix, prefix)
	}

	return nil
}
