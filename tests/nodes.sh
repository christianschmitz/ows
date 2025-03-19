. ./assert.sh

NODE_PID=""

# Don't run in subshell, so trap is triggered by top-level exit
start_node() {
    local node_private_key=$1
    local initial_config=$2
    
    local log_file=$(node_log_file)

    OWS_PRIVATE_KEY=$node_private_key \
    OWS_INITIAL_CONFIG=$initial_config \
    nohup ../dist/ows-node --test-dir $TEST_DIR > $log_file 2>&1 &

    NODE_PID=$!

    trap_add "stop_node $NODE_PID" EXIT
}

stop_node() {
    local pid=$1

    if ps -p $pid > /dev/null; then
        kill -9 $pid
    fi
}

node_log_file() {
    echo "${TEST_DIR}/nodes.log"
}