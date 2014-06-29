#!/bin/bash
set -e

if [ -z "$1" ] ; then
    echo "Must provide github release"
    exit 1
fi

RELEASE="$1"
TAG=${RELEASE//v/}

cd $(dirname $0)
docker build -t="lox24/package-proxy" .

docker tag lox24/package-proxy lox24/package-proxy:$TAG
docker push lox24/package-proxy:$TAG

docker tag lox24/package-proxy lox24/package-proxy:latest
docker push lox24/package-proxy:latest