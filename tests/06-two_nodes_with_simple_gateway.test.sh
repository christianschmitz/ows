. keys.sh
. nodes.sh
. projects.sh
. resources.sh

TEST_NAME="06-Two nodes with simple gateway"

init_test_dir

test() {
    # we must use different ports for different nodes, because during this test they all run on the same machine
    local node1_api_port=9000
    local node1_gossip_port=9001
    local node2_api_port=9002
    local node2_gossip_port=9003

    local gateway_port=8080
    local node1_gateway_port=$gateway_port
    local node2_port_offset=10
    local node2_gateway_port=$(($node1_gateway_port + $node2_port_offset))

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

    # 4. Start the node, and give it some time
    start_node $node1_private_key $project
    sleep 3

    # 5. Create a simple CommonJS function
    local handler_path="${TEST_DIR}/index.cjs"
    echo 'module.exports = function handler() {return "Hello World"};' > $handler_path
    local asset_id=$(upload_asset $client $project $handler_path)
    local function_id=$(add_function $client $project $asset_id)

    # 6. Add a second node to the ledger
    local node_id=$(add_node $client $project $node2_public_key $node2_api_port $node2_gossip_port)
    sleep 1

    # 7. Start the second node, and give it some time
    start_node $node2_private_key $project $node2_port_offset
    sleep 2

    # 8. Add the gateway and the endpoint
    local gateway=$(add_gateway $client $project $gateway_port)
    add_gateway_endpoint $client $project $gateway "GET" "/" $function_id

    # 11. Query the newly created endpoint
    response1=$(curl -sS -o - "http://127.0.0.1:${node1_gateway_port}/")
    response2=$(curl -sS -o - "http://127.0.0.1:${node2_gateway_port}/")

    assert_equals "$response1" '"Hello World"' \
        "gateway endpoint on node1 responds correctly"
    assert_equals "$response2" '"Hello World"' \
        "gateway endpoint on node2 responds correctly"
}

test