#!/bin/bash
ows_containers=$(docker ps --filter "name=ows_nodejs_containernode1" -a -q)

if [[ "$ows_containers" != "" ]]; then
    docker stop $ows_containers
    docker rm $ows_containers
    docker system prune -f
fi