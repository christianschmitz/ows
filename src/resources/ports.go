package resources

//go:generate go run IANAPorts.go

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"slices"

	"ows/ledger"
)

var PopularPorts = []ledger.Port{
	8080,
}

const maxPortGenIter = 1000

func RandomPort() (ledger.Port, error) {
	var p ledger.Port
	isAvailable := false
	iter := 0

	for !isAvailable {
		if iter >= maxPortGenIter {
			return p, errors.New("too many iterations")
		}

		if err := binary.Read(rand.Reader, binary.LittleEndian, &p); err != nil {
			return p, err
		}

		if !isIANAPort(p) && !isPopularPort(p) {
			isAvailable = true
		}

		iter++
	}

	return p, nil
}

func isIANAPort(p ledger.Port) bool {
	return slices.Contains(IANAPorts, uint16(p))
}

func isPopularPort(p ledger.Port) bool {
	return slices.Contains(PopularPorts, p)
}
