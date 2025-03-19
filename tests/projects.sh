. ./assert.sh

TEST_DIR=""

# Don't run in a subshell, so trap is triggered by top-level EXIT
init_test_dir() {
    TEST_DIR=$(mktemp -d)

    trap_add "rm -fr $TEST_DIR" EXIT
}

new_project() {
    local client_private_key=$1
    local node_public_key=$2
    local node_api_port=$3
    local node_gossip_port=$4

    OWS_PRIVATE_KEY=$client_private_key \
    ../dist/ows projects new \
        test "$node_public_key" 127.0.0.1 \
        --api-port $node_api_port \
        --gossip-port $node_gossip_port \
        --test-dir $TEST_DIR
}

get_project_id() {
    echo $2
}

get_project_initial_config() {
    echo $4
}