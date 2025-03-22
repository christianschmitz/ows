#!/bin/bash

TEST_DIR="./tests"

echo "Running integration tests in $TEST_DIR"

cd $TEST_DIR

tests=$(ls ./*.test.sh)
failure=0

for test_script in $tests; do
    bash $test_script
    code=$?

    if [[ code -ne 0 ]]; then
        failure=$code
    fi
done

echo "Cleaning up"

cd ../
./clean.sh

if [[ failure -ne 0 ]]; then
    echo "TESTS FAILED, SEE ABOVE" >&2
    exit $failure
fi