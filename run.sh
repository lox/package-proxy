#!/bin/bash
set -e

git_version() {
  git describe --tags --abbrev=4 --dirty --always
}

cd $GOPATH/src/github.com/lox/package-proxy
go build -ldflags "-X main.version $(git_version)" .
./package-proxy -version

exec "$@"
