# Generates a random ed25519 key pair.
#
# To get the private key and public key as separate variables, use:
# ```sh
# key_pair=$(gen_key_pair)
# private_key=$(get_private_key $key_pair)
# public_key=$(get_public_key $key_pair)
# ```
gen_key_pair() {
    ../dist/ows key gen
}

get_private_key() {
    echo $2
}

get_public_key() {
    echo $4
}