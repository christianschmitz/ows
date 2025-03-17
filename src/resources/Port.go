package resources

//go:generate go run IANAPorts.go

import (
	"crypto/rand"
	"encoding/binary"
	"slices"
)

// convenvience wrapper for uint16
type Port uint16

var PopularPorts = []Port{
	8080,
}

func RandomPort() (Port, error) {
	var p Port
	isAvailable := false

	for !isAvailable {
		if err := binary.Read(rand.Reader, binary.LittleEndian, &p); err != nil {
			return p, err
		}

		if !slices.Contains(IANAPorts, uint16(p)) && !slices.Contains(PopularPorts, p) {
			isAvailable = true
		}
	}

	return p, nil
}
