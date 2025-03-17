package main

import (
	"encoding/base64"
	"errors"
	"log"
	"strings"

	"github.com/btcsuite/btcutil/bech32"
	"golang.org/x/crypto/blake2b"
)

func StringifyHumanReadableBytes(prefix string, bs []byte) string {
	conv, err := bech32.ConvertBits(bs, 8, 5, true)
	if err != nil {
		log.Fatal(err)
	}

	s, err := bech32.Encode(prefix, conv)
	if err != nil {
		log.Fatal(err)
	}

	return s
}

func EncodeBytes(bs []byte) string {
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(bs)
}

func ParseHumanReadableBytes(s string, expectedPrefix string) ([]byte, error) {
	s = strings.TrimSpace(s)

	prefix, bs, err := bech32.Decode(s)

	if err != nil {
		return nil, err
	}

	if expectedPrefix != "*" && prefix != expectedPrefix {
		return nil, errors.New("unexpected bech32 prefix " + prefix)
	}

	return bech32.ConvertBits(bs, 5, 8, false)
}

func DecodeBytes(s string) ([]byte, error) {
	s = strings.TrimSpace(s)

	return base64.URLEncoding.WithPadding(base64.NoPadding).DecodeString(s)
}

func DigestCompact(bs []byte) []byte {
	hasher, err := blake2b.New(16, nil)
	if err != nil {
		log.Fatal(err)
	}

	hasher.Write(bs)
	hash := hasher.Sum(nil)

	return hash
}
