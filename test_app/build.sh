#!/bin/bash

echo "Example build script to build the binary via docker-machine. You should already have docker-machine setup (with a default docker-machine) for this to work."

docker-machine start default
eval "$(docker-machine env default)"

set -e

docker build --tag test_app:1 .
