# Fails with error code 1 if expected isn't equal to actual, ignoring
# whitespace.
assert_equals() {
    local actual=$1
    local expected=$2
    local message=$3

    local actual_clean=$(remove_whitespace $actual)
    local expected_clean=$(remove_whitespace $expected)

    if [[ "$actual_clean" != "$expected_clean" ]]; then
        echo "[$TEST_NAME] $message [FAIL]"
        echo "[$TEST_NAME]   expected='$expected', actual='$actual'"
        exit 1
    else
        echo "[$TEST_NAME] $message [OK]"
    fi  
}

# Ignores the prefix
assert_bech32_equals() {
    local actual=$1
    local expected=$2
    local message=$3

    local actual_clean=$(get_bech32_payload $actual)
    local expected_clean=$(get_bech32_payload $expected)

    if [[ "$actual_clean" != "$expected_clean" ]]; then
        echo "[$TEST_NAME] $message [FAIL]"
        echo "[$TEST_NAME]   expected='$expected', actual='$actual'"
        exit 1
    else
        echo "[$TEST_NAME] $message [OK]"
    fi
}

get_bech32_payload() {
    local id=$(remove_whitespace $1)
    local prefix=$(get_bech32_prefix $id)

    start=$(get_string_length $prefix)

    echo ${id:start:27}
}

get_bech32_prefix() {
    local id=$(remove_whitespace $1)

    echo $id | sed "s/1.*//"
}

get_string_length() {
    local s=$1

    echo ${#s}
}

# Removes surrounding and internal spaces.
# Doesn't remove newlines.
remove_whitespace() {
    echo "$*" | tr -d '[:space:]'
}



# From: https://stackoverflow.com/questions/3338030/multiple-bash-traps-for-the-same-signal
# Appends a command to a trap
#
# - 1st arg:  code to add
# - remaining args:  names of traps to modify
#
trap_add() {
    trap_add_cmd=$1; shift || fatal "${FUNCNAME} usage error"
    for trap_add_name in "$@"; do
        trap -- "$(
            # helper fn to get existing trap command from output
            # of trap -p
            extract_trap_cmd() { printf '%s\n' "$3"; }
            # print existing trap command with newline
            eval "extract_trap_cmd $(trap -p "${trap_add_name}")"
            # print the new trap command
            printf '%s\n' "${trap_add_cmd}"
        )" "${trap_add_name}" \
            || fatal "unable to add to trap ${trap_add_name}"
    done
}
# set the trace attribute for the above function.  this is
# required to modify DEBUG or RETURN traps because functions don't
# inherit them unless the trace attribute is set
declare -f -t trap_add

fatal() { error "$@"; exit 1; }