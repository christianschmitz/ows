# Add a nodejs CommonJS function
add_function() {
    local client_private_key=$1
    local initial_config=$2
    local asset_id=$3

    OWS_PRIVATE_KEY=$client_private_key \
    OWS_INITIAL_CONFIG=$initial_config \
    ../dist/ows \
        functions add nodejs $asset_id \
        --test-dir $TEST_DIR
}

add_gateway() {
    local client_private_key=$1
    local initial_config=$2
    local port=$3

    OWS_PRIVATE_KEY=$client_private_key \
    OWS_INITIAL_CONFIG=$initial_config \
    ../dist/ows \
        gateways add $port \
        --test-dir $TEST_DIR
}

add_gateway_endpoint() {
    local client_private_key=$1
    local initial_config=$2
    local gateway=$3
    local method=$4
    local path=$5
    local function_id=$6

    OWS_PRIVATE_KEY=$client_private_key \
    OWS_INITIAL_CONFIG=$initial_config \
    ../dist/ows \
        gateways endpoints add $gateway $method $path $function_id \
        --test-dir $TEST_DIR
}

list_gateways() {
    local client_private_key=$1
    local initial_config=$2

    OWS_PRIVATE_KEY=$client_private_key \
    OWS_INITIAL_CONFIG=$initial_config \
    ../dist/ows \
        gateways list \
        --only-ids \
        --test-dir $TEST_DIR
}

# Upload a single asset, echoing the asset id
upload_asset() {
    local client_private_key=$1
    local initial_config=$2
    local path=$3

    output=$(
        OWS_PRIVATE_KEY=$client_private_key \
        OWS_INITIAL_CONFIG=$initial_config \
        ../dist/ows \
            assets upload $path \
            --test-dir $TEST_DIR
    )

    echo $output | awk '{print $2}'
}