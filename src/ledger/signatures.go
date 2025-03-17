package ledger

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/fxamacker/cbor/v2"
	"github.com/tyler-smith/go-bip39"
)

// When a new client or node is started its private key is passed via an
// environment variable with the following name:
const PrivateKeyEnvName = "OWS_PRIVATE_KEY"

// Although an Ed25519 private key is always 64 bytes, and an Ed25519 public key is always 32 bytes, it is more convenient
//  to reuse the format from the crypto/ed25519 package (i.e. use []byte instead of [64]byte or [32]byte).
type PrivateKey ed25519.PrivateKey
type PublicKey ed25519.PublicKey

type KeyPair struct {
	Private PrivateKey `cbor:"0,keyasint"`
	Public  PublicKey  `cbor:"1,keyasint"`
}

type Signature struct {
	Key   PublicKey   `cbor:"0,keyasint"`
	Bytes [64]byte `cbor:"1,keyasint"`
}

func DecodeKeyPair(bytes []byte) (*KeyPair, error) {
	var p KeyPair

	if err := cbor.Unmarshal(bytes, &p); err != nil {
		return nil, err
	}

	if err := p.Validate(); err != nil {
		return nil, err
	}

	return &p, nil
}

func RandomKeyPair() (*KeyPair, error) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	if len(publicKey) != 32 {
		return nil, errors.New("generated public key not exactly 32 bytes long")
	}

	if len(privateKey) != 64 {
		return nil, errors.New("generated private key not exactly 64 bytes long")
	}

	return &KeyPair{
		PrivateKey(privateKey),
		PublicKey(publicKey),
	}, nil
}

func ReadKeyPair(path string) (*KeyPair, error) {
	// file exists
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return DecodeKeyPair(bytes)
}

func RestoreKeyPair(phrase []string) (*KeyPair, error) {
	if len(phrase) != 24 {
		return nil, fmt.Errorf("expected phrase with 24 words, got %d words", len(phrase))
	}

	seed, err := bip39.EntropyFromMnemonic(strings.Join(phrase, " "))
	if err != nil {
		return nil, fmt.Errorf("invalid phrase 0-24 (%v)", err)
	}

	privateKey := PrivateKey(ed25519.NewKeyFromSeed(seed))

	return privateKey.KeyPair(), nil
}

func (p *KeyPair) Encode() ([]byte, error) {
	return cbor.Marshal(*p)
}

func (p *KeyPair) Phrase() ([]string, error) {
	bs := []byte(p.Private)

	// Only the first 32 bytes contain unique information (the latter 32 bytes are derived from the former 32 bytes)
	seed := bs[0:32]	

	phrase, err := bip39.NewMnemonic(seed)
	if err != nil {
		return nil, err
	}

	return strings.Split(phrase, " "), nil
}

func (p *KeyPair) Sign(message []byte) (Signature, error) {
	rawSigBytes := ed25519.Sign(ed25519.PrivateKey(p.Private), message)

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
	message := cs.withoutSignatures().Encode()

	return p.Sign(message)
}

func (p *KeyPair) Validate() error {
	check := ed25519.PrivateKey(p.Private).Public()

	if bs, ok := check.(ed25519.PublicKey); ok {
		if !bytes.Equal(p.Public, bs) {
			return errors.New("invalid key pair, public key doesn't correspond to private key")
		}
	} else {
		return fmt.Errorf("expected a bytestring public key, but got a %T", check)
	}

	return nil
}

func (p *KeyPair) Write(path string) error {
	bytes, err := p.Encode()
	if err != nil {
		return err
	}

	return WriteSafe(path, bytes)
}

func ParsePrivateKey(s string) (PrivateKey, error) {
	seed, err := parseKeyBytes(s)
	if err != nil {
		return nil, fmt.Errorf("invalid PrivateKey (%v)", err)
	}

	return PrivateKey(ed25519.NewKeyFromSeed(seed)), nil
}

func ParsePublicKey(s string) (PublicKey, error) {
	bs, err := parseKeyBytes(s)
	if err != nil {
		return nil, fmt.Errorf("invalid PubKey (%v)", err)
	}

	return bs, nil
}

func parseKeyBytes(s string) ([]byte, error) {
	s = strings.TrimSpace(s)

	var key []byte
	var err error
	if len(s) == 64 {
		key, err = hex.DecodeString(s)
	} else {
		// try base64
		key, err = base64.URLEncoding.WithPadding(base64.NoPadding).DecodeString(s)
	}

	if err != nil {
		return nil, err
	}

	if len(key) != 32 {
		return nil, fmt.Errorf("expected a 32 byte key, got %d bytes", len(key))
	}

	return key, nil
}

func (k PrivateKey) KeyPair() *KeyPair {
	pubKey := ed25519.PrivateKey(k).Public()

	if bs, ok := pubKey.(ed25519.PublicKey); ok {
		return &KeyPair{
			Private: k,
			Public: PublicKey(bs),
		}
	} else {
		panic("derived public key isn't a bytestring")
	}
}

func (k PrivateKey) String() string {
	seed := []byte(k)[0:32]

	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(seed)
} 

func (s Signature) Verify(message []byte) bool {
	return ed25519.Verify([]byte(s.Key)[:], message, s.Bytes[:])
}

// Validates all signatures and then collects unique signers.
func (cs *ChangeSet) validateSignatures() ([]PublicKey, error) {
	users := []PublicKey{}

	message := cs.withoutSignatures().Encode()

	for i, s := range cs.Signatures {
		if !s.Verify(message) {
			return nil, fmt.Errorf("invalid signature %d", i)
		}

		unique := true

		for _, user := range users {
			if bytes.Equal(s.Key, user) {
				unique = false
				break
			}
		}

		if unique {
			users = append(users, s.Key)
		}
	}

	return users, nil
}

// Creates a copy without signatures, which is 
func (cs *ChangeSet) withoutSignatures() *ChangeSet {
	return &ChangeSet{
		Prev: cs.Prev,
		Actions: cs.Actions,
		Signatures: []Signature{},
	}
}