. keys.sh
. nodes.sh
. projects.sh
. resources.sh

TEST_NAME="04-Add second node"

init_test_dir

test() {
    # we must use different ports for different nodes, because during this test they all run on the same machine
    local node1_api_port=9000
    local node1_gossip_port=9001
    local node2_api_port=9002
    local node2_gossip_port=9003

    # 1. Generate the client key pair
    local client_key_pair=$(gen_key_pair)
    local client=$(get_private_key $client_key_pair)

    # 2. Generate the node key pairs
    local node1_key_pair=$(gen_key_pair)
    local node1_private_key=$(get_private_key $node1_key_pair)
    local node1_public_key=$(get_public_key $node1_key_pair)
    local node2_key_pair=$(gen_key_pair)
    local node2_private_key=$(get_private_key $node2_key_pair)
    local node2_public_key=$(get_public_key $node2_key_pair)

    # 3. Create the initial project config, using the first node for bootstrapping
    local project=$(get_project_initial_config $(new_project $client $node1_public_key $node1_api_port $node1_gossip_port))

    # 4. Start the node
    start_node $node1_private_key $project

    # 5. Give the node some time to start up the API
    sleep 2

    # 6. Add a second node to the ledger
    local node_id=$(add_node $client $project $node2_public_key $node2_api_port $node2_gossip_port)

    # 7. Start the second node
    start_node $node2_private_key $project

    # 8. Give the second node some time to start up
    sleep 1

    # assert that two nodes are running
    # TODO: turn this into a bash function, and move into ./assert.sh
    if [[ $(ps -e | grep "ows-node" | wc -l) -ne 2 ]]; then
        echo "[$TEST_NAME] two ows-node processes running [FAIL]"
        exit 1
    else
        echo "[$TEST_NAME] two ows-node processes running [OK]"
    fi
}

test