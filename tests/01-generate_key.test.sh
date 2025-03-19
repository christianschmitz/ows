. assert.sh
. keys.sh

TEST_NAME="01-Generate Key"

# Test that the `ows key gen` command prints a key pair to stdout in the
# correct format.
test() {
    local key_pair=$(gen_key_pair)
    local private_key=$(get_private_key $key_pair)
    local public_key=$(get_public_key $key_pair)

    assert_equals "PrivateKey: $private_key PublicKey: $public_key" "$key_pair" "key pair stdout format"
}

test