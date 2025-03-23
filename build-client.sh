#!/bin/bash

VERSION=$(git describe --tags --abbrev=0)

cd ./src
go build -ldflags "-X main.Version=$VERSION" -o ../dist/ows ./client/*.go