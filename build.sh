#!/bin/bash

echo "Example build script to build the binary via docker-machine. You should already have docker-machine setup (with a default docker-machine) for this to work."

docker-machine start default
eval "$(docker-machine env default)"

set -e

docker build --tag secret_squirrel:dont_push_me .
docker run -v /tmp:/tmp/export -it secret_squirrel:dont_push_me cp /tmp/secret_squirrel /tmp/export/secret_squirrel
docker run -v /tmp:/tmp/export -it secret_squirrel:dont_push_me cp /tmp/secret_squirrel_s3 /tmp/export/secret_squirrel_s3
docker-machine ssh default cat /tmp/secret_squirrel | cat > lib/secret_squirrel
docker-machine ssh default cat /tmp/secret_squirrel_s3 | cat > lib/secret_squirrel_s3

