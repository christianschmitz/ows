ledger_summary() {
    local client_private_key=$1
    local initial_config=$2

    OWS_PRIVATE_KEY=$client_private_key \
    OWS_INITIAL_CONFIG=$initial_config \
    ../dist/ows \
        ledger list \
        --test-dir $TEST_DIR
}