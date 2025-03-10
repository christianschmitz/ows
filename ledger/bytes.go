package ledger

import "encoding/base64"

// TODO: bech32 with custom prefix
func StringifyBytes(bytes []byte) string {
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(bytes[:])	
}

// TODO: bech32 with custom prefix
func ParseBytes(str string) ([]byte, error) {
	return base64.URLEncoding.WithPadding(base64.NoPadding).DecodeString(str)
}