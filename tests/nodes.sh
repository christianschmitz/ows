. ./assert.sh

NODE_PID=""
NODE_COUNT=0

# Add node with address 127.0.0.1
add_node() {
    local client_private_key=$1
    local initial_config=$2
    local node_public_key=$3
    local api_port=$4
    local gossip_port=$5

    OWS_PRIVATE_KEY=$client_private_key \
    OWS_INITIAL_CONFIG=$initial_config \
    ../dist/ows \
        nodes add "$node_public_key" 127.0.0.1 \
        --api-port $api_port \
        --gossip-port $gossip_port \
        --test-dir $TEST_DIR
}

# Don't run in subshell, so trap is triggered by top-level exit
start_node() {
    local node_private_key=$1
    local initial_config=$2
    
    local log_file=$(node_log_file)

    OWS_PRIVATE_KEY=$node_private_key \
    OWS_INITIAL_CONFIG=$initial_config \
    nohup ../dist/ows-node --test-dir $TEST_DIR &>> $log_file &

    NODE_COUNT=$((NODE_COUNT + 1))
    NODE_PID=$!

    trap_add "stop_node $NODE_PID" EXIT
}

stop_node() {
    local pid=$1

    if ps -p $pid > /dev/null; then
        kill $pid # don't call with -9, so the nodes have time to clean up, and to avoid an ugly kill message being printed to stderr
    fi
}

node_log_file() {
    echo "${TEST_DIR}/node${NODE_COUNT}.log"
}