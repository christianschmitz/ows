#!/bin/bash
go -C ./src mod tidy
./build-client.sh
./build-node.sh
./build-doc.sh
