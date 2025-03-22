. keys.sh
. nodes.sh
. projects.sh
. resources.sh

TEST_NAME="03-Single node with simple gateway"

init_test_dir

# TODO: reuse docker container from previous runs to speed up this test
test() {
    local node_api_port=9000
    local node_gossip_port=9001
    local gateway_port=8080

    # 1. Generate the client key pair
    local client_key_pair=$(gen_key_pair)
    local client=$(get_private_key $client_key_pair)

    # 2. Generate the node key pair
    local node_key_pair=$(gen_key_pair)
    local node_private_key=$(get_private_key $node_key_pair)
    local node_public_key=$(get_public_key $node_key_pair)

    # 3. Create the initial project config
    local project=$(get_project_initial_config $(new_project $client $node_public_key $node_api_port $node_gossip_port))

    # 4. Start the node
    start_node $node_private_key $project

    # 5. Give the node some time to start up the API
    sleep 1

    # 6. Create a gateway at port 8080
    local gateway=$(add_gateway $client $project $gateway_port)

    # 7. List all gateways
    local all_gateways=$(list_gateways $client $project)

    assert_equals $gateway $all_gateways "a single gateway is created"

    # 8. Create a simple CommonJS function
    local handler_path="${TEST_DIR}/index.cjs"
    echo 'module.exports = function handler() {return "Hello World"};' > $handler_path
    local asset_id=$(upload_asset $client $project $handler_path)
    local function_id=$(add_function $client $project $asset_id)

    # 9. Add the function as an endpoint to the gateway
    add_gateway_endpoint $client $project $gateway "GET" "/" $function_id

    # 10. Give the node some time to initialize the nodejs docker container
    sleep 2

    # 11. Query the newly created endpoint
    response=$(curl -sS -o - "http://127.0.0.1:${gateway_port}/")

    assert_equals "$response" '"Hello World"' \
        "gateway endpoint responds correctly" 
}

test