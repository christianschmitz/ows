package ledger

import (
	"encoding/base64"
	"errors"
	"log"
	"github.com/btcsuite/btcutil/bech32"
)

func StringifyHumanReadableBytes(prefix string, bs []byte) string {
	conv, err := bech32.ConvertBits(bs, 8, 5, true)
	if err != nil {
		log.Fatal(err)
	}

	str, err := bech32.Encode(prefix, conv)
	if err != nil {		
		log.Fatal(err)
	}

	return str
}

func StringifyCompactBytes(bs []byte) string {
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(bs)
}

func ParseHumanReadableBytes(str string, expectedPrefix string) ([]byte, error) {
	prefix, bs, err := bech32.Decode(str)

	if err != nil {
		return nil, err
	}

	if prefix != expectedPrefix {
		return nil, errors.New("unexpected bech32 prefix " + prefix)
	}

	return bech32.ConvertBits(bs, 5, 8, false)
}

func ParseCompactBytes(str string) ([]byte, error) {
	return base64.URLEncoding.WithPadding(base64.NoPadding).DecodeString(str)
}