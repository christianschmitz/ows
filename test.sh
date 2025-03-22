#!/bin/bash

TEST_DIR="./tests"

echo "Running integration tests in $TEST_DIR"

cd $TEST_DIR

tests=$(ls ./*.test.sh)

for test_script in $tests; do
    bash $test_script
done

cd ../
./clean.sh