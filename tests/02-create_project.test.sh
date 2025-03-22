. assert.sh
. keys.sh
. ledger.sh
. nodes.sh
. projects.sh

TEST_NAME="02-Create Project"

init_test_dir

# Test the creation of a new project with a single node.
test() {
    local node_api_port=9000
    local node_gossip_port=9001

    # 1. Generate the client key pair
    local client_key_pair=$(gen_key_pair)
    local client_private_key=$(get_private_key $client_key_pair)

    # 2. Generate the node key pair
    local node_key_pair=$(gen_key_pair)
    local node_private_key=$(get_private_key $node_key_pair)
    local node_public_key=$(get_public_key $node_key_pair)

    # 3. Create the initial project config
    local project=$(new_project $client_private_key $node_public_key $node_api_port $node_gossip_port)
    local project_id=$(get_project_id $project)
    local initial_config=$(get_project_initial_config $project)

    # 4. Start the node
    start_node $node_private_key $initial_config
    local node_pid=$NODE_PID # node pid is stored in global variable

    # 5. Give the node some time to start up the API
    sleep 1

    # 6. Sync the client with the node and return the change set IDs
    local id_chain=$(ledger_summary $client_private_key $initial_config)

    assert_bech32_equals $project_id $id_chain \
        "project id bech32 payload equals first change set id bech32 payload"
}

test
