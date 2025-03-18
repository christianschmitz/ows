package ledger

import (
	"errors"
	"fmt"
)

// A helper class to detect forks and sync ledgers.
type ChangeSetIDChain struct {
	IDs []ChangeSetID
}

// Find a change set with the given id. Searches from the end to the beginning
// of `l.Changes`.
func (l *Ledger) FindChange(id ChangeSetID) (*ChangeSet, bool) {
	n := len(l.Changes)

	for i := n; i >= 0; i-- {
		if i > 0 && i == n {
			if id == l.Head() {
				return &(l.Changes[i-1]), true
			}
		} else if i < n {
			if i > 1 && id == l.Changes[i].Prev {
				return &(l.Changes[i-1]), true
			}
		} else if i > 0 {
			if id == l.Changes[i].ID() {
				return &(l.Changes[i]), true
			}
		}
	}

	return nil, false
}

// Returns all the change set ids.
func (l *Ledger) IDChain() *ChangeSetIDChain {
	n := len(l.Changes)
	ids := make([]ChangeSetID, n)

	for i := 0; i < n; i++ {
		if i+1 == n {
			ids[i] = l.Head()
		} else if i+1 < n {
			ids[i] = l.Changes[i+1].Prev
		} else {
			ids[i] = l.Changes[i].ID()
		}
	}

	return &ChangeSetIDChain{
		ids,
	}
}

// Removes [until+1:] changes, and revalidates from the beginning.
func (l *Ledger) Keep(until int) {
	l.Changes = l.Changes[0 : until+1]

	if err := l.Validate(); err != nil {
		panic(fmt.Errorf("old state of ledger is invalid (%v)", err))
	}
}

// Returns the index of the latest common change set.
// Returns an error if no intersection is found.
func (ca *ChangeSetIDChain) Intersect(cb *ChangeSetIDChain) (int, error) {
	n := min(len(ca.IDs), len(cb.IDs))

	for i := 0; i < n; i++ {
		a := ca.IDs[i]
		b := cb.IDs[i]

		if a == b {
			continue
		} else {
			if i == 0 {
				return 0, errors.New("no common intersection found")
			} else {
				return i - 1, nil
			}
		}
	}

	return n - 1, nil
}
