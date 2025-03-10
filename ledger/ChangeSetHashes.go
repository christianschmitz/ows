package ledger

import (
	"encoding/json"
	"errors"
	"log"
)

type ChangeSetHashes struct {
	Hashes []ChangeSetHash
}

// convert to JSON list of strings
func (c *ChangeSetHashes) Stringify() string {
	rawHashes := make([]string, len(c.Hashes))

	for i, h := range c.Hashes {
		rawHashes[i] = StringifyChangeSetHash(h)
	}

	j, err := json.Marshal(rawHashes)
	if err != nil {
		log.Fatal(err)
	}

	return string(j)
}

// returns an error if no intersection is found
// the returned int is the index of the intersection in `a` and `b`
func (a *ChangeSetHashes) FindIntersection(b *ChangeSetHashes) (int, error) {
	n := len(a.Hashes)
	if (n > len(b.Hashes)) {
		n = len(b.Hashes)
	}

	for i := 0; i < n; i++ {
		ha := a.Hashes[i]
		hb := b.Hashes[i]

		if IsSameChangeSetHash(ha, hb) {
			continue
		} else {
			if (i == 0) {
				return -1, errors.New("no common intersection found")
			} else {
				return i-1, nil
			}
		}
	}

	return n-1, nil
}
