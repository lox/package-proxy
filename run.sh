#!/bin/bash
set -e

git_version() {
  git describe --tags --abbrev=4 --dirty --always
}

cd $GOPATH/src/github.com/lox/package-proxy

go install -ldflags "-X main.version $(git_version)" .

exec "$@"
