package ledger

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"github.com/fxamacker/cbor/v2"
	"os"
)

type PrivateKey = [64]byte

type PubKey = [32]byte

type KeyPair struct {
	Private PrivateKey `cbor:"0,keyasint"`
	Public  PubKey     `cbor:"1,keyasint"`
}

func StringifyPubKey(key PubKey) string {
	return hex.EncodeToString(key[:])
}

func DecodeKeyPair(bytes []byte) (*KeyPair, error) {
	var p KeyPair

	err := cbor.Unmarshal(bytes, &p)

	if err != nil {
		return nil, err
	}

	return &p, nil
}

func GenerateKeyPair() (*KeyPair, error) {
	rawPublicKey, rawPrivateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	if len(rawPublicKey) != 32 {
		return nil, errors.New("generated public key not exactly 32 bytes long")
	}

	if len(rawPrivateKey) != 64 {
		return nil, errors.New("generated private key not exactly 64 bytes long")
	}

	publicKey := [32]byte{}
	privateKey := [64]byte{}

	for i, b := range rawPublicKey {
		publicKey[i] = b
	}

	for i, b := range rawPrivateKey {
		privateKey[i] = b
	}

	return &KeyPair{
		privateKey,
		publicKey,
	}, nil
}

func ReadKeyPair(path string, generateNewIfNotExists bool) (*KeyPair, error) {
	if _, err := os.Stat(path); err == nil {
		bytes, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}

		return DecodeKeyPair(bytes)
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	} else if generateNewIfNotExists {
		pair, err := GenerateKeyPair()

		if err != nil {
			return nil, err
		}

		err = pair.Write(path)

		if err != nil {
			return nil, err
		}

		return pair, nil
	} else {
		return nil, err
	}
}

func (p *KeyPair) Encode() ([]byte, error) {
	return cbor.Marshal(*p)
}

func (p *KeyPair) Sign(message []byte) (Signature, error) {
	rawSigBytes := ed25519.Sign(p.Private[:], message)

	if len(rawSigBytes) != 64 {
		return Signature{}, errors.New("ed25519 signature not exactly 64 bytes long")
	}

	sigBytes := [64]byte{}

	for i, b := range rawSigBytes {
		sigBytes[i] = b
	}

	return Signature{p.Public, sigBytes}, nil
}

func (p *KeyPair) SignChangeSet(cs *ChangeSet) (Signature, error) {
	message, err := cs.Encode(true)
	if err != nil {
		return Signature{}, err
	}

	return p.Sign(message)
}

func (p *KeyPair) Write(path string) error {
	bytes, err := p.Encode()

	if err != nil {
		return err
	}

	return os.WriteFile(path, bytes, 0644)
}
