#!/bin/bash
set -e
set -x

export GITHUB_USER=lox
export GITHUB_REPO=package-proxy

git_version() {
  git describe --tags --abbrev=4 --dirty --always
}

VERSION="$1"
NAME="package-proxy-${2:-linux-amd64}"
FILE="${3:-$GOBIN/package-proxy}"

if [ ! -f $FILE ] ; then
  go build -ldflags "-X main.version $VERSION" -o $FILE .
fi

if [ -z "$VERSION" ] ; then
  VERSION=$(git_version)
fi

if ! github-release info -t "$VERSION" &>/dev/null ; then
  echo "creating release $VERSION"
  github-release release -t "$VERSION"
fi

echo "uploading $FILE => $NAME for $VERSION"
github-release upload -t "$VERSION" --file "$FILE" --name "$NAME"