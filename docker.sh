#!/bin/bash

PACKAGE=github.com/lox/package-proxy

while [[ $1 == -* ]] ; do
  case "$1" in
    --build)        build=1   ;;
    --run)          run=1     ;;
    --dev)          dev=1     ;;
    --build-linux)  buildlinux=1 ;;
  esac
  shift
done

if [ -n "$build" ] ; then
  docker build --tag="lox24/package-proxy" .
fi

if [ -n "$run" ] ; then
  docker rm -f package-proxy &>/dev/null || true
  docker run \
    --tty --interactive --rm \
    --name package-proxy \
    --publish 3142:3142 \
    --publish 3143:3143 \
    --volume /tmp/vagrant-cache/generic/:/tmp/cache \
    lox24/package-proxy:latest "$@"
fi

if [ -n "$dev" ] ; then
  docker rm -f package-proxy &>/dev/null || true
  docker run \
    --tty --interactive --rm \
    --name package-proxy \
    --publish 3142:3142 \
    --publish 3143:3143 \
    --env GITHUB_TOKEN=$GITHUB_TOKEN \
    --volume /projects/go/src/$PACKAGE:/go/src/$PACKAGE \
    --volume /tmp/vagrant-cache/generic/:/tmp/cache \
    --entrypoint "/run.sh" \
    lox24/package-proxy:latest "$@"
fi

if [ -n "$buildlinux" ] ; then
  docker rm -f package-proxy &>/dev/null || true
  CID=$(docker run \
    --detach  \
    --name package-proxy \
    --publish 3142:3142 \
    --publish 3143:3143 \
    --env GITHUB_TOKEN=$GITHUB_TOKEN \
    --volume /projects/go/src/$PACKAGE:/go/src/$PACKAGE \
    --volume /tmp/vagrant-cache/generic/:/tmp/cache \
    --entrypoint "/run.sh" \
    lox24/package-proxy true)

  # https://github.com/dotcloud/docker/issues/3986
  docker cp $CID:/go/bin/package-proxy $CID &>/dev/null || true

  mv $CID/package-proxy package-proxy-linux-amd64
  rm -rf $CID/
  file package-proxy-linux-amd64
fi