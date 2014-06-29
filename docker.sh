#!/bin/bash
set -e

cd $(dirname $0)

PACKAGE=github.com/lox/package-proxy

while [[ $1 == -* ]] ; do
  case "$1" in
    --build)        build=1   ;;
    --run)          run=1     ;;
    --runit)        runit=1   ;;
    --dev)          dev=1     ;;
    --build-linux)  buildlinux=1 ;;
  esac
  shift
done

if [ -n "$build" ] ; then
  echo "Building development docker container"
  docker build --tag="lox24/package-proxy" .
fi

if [ -n "$dev" ] ; then
  echo "Running development docker container"
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
    lox24/package-proxy "$@"
fi

if [ -n "$buildlinux" ] ; then
  trap "rm .cidfile" EXIT
  test -f package-proxy && rm package-proxy

  echo "Building linux build container"
  docker rm -f package-proxy &>/dev/null || true
  docker build --tag="lox24/package-proxy" .
  docker run \
    --interactive --tty --rm \
    --publish 3142:3142 \
    --publish 3143:3143 \
    --env GITHUB_TOKEN=$GITHUB_TOKEN \
    --volume /projects/go/src/$PACKAGE:/go/src/$PACKAGE \
    --volume /tmp/vagrant-cache/generic/:/tmp/cache \
    --entrypoint "/run.sh" \
    --cidfile=".cidfile" \
    lox24/package-proxy true

  echo "Building release docker container (with linux-amd64 binary)"
  mv package-proxy release/package-proxy-linux-amd64
  docker build -t="lox24/package-proxy" ./release
fi

if [[ -n "$run" || -n "$runit" ]] ; then
  if [ -n "$runit" ] ; then
    args="--interactive --tty --rm"
  else
    args="--detach"
  fi

  docker rm -f package-proxy &>/dev/null || true
  docker run \
    $args \
    --name package-proxy \
    --publish 3142:3142 \
    --publish 3143:3143 \
    --volume /tmp/vagrant-cache/generic/:/tmp/cache \
    lox24/package-proxy "$@"
fi