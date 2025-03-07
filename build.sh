#!/bin/bash
go mod tidy
./build-client.sh
./build-node.sh
./build-doc.sh
