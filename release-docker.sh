#!/bin/bash
set -e

if [ -z "$1" ] ; then
    echo "Must provide github release"
    exit 1
fi

RELEASE="$1"
TAG=${RELEASE//v/}
URL="https://github.com/lox/package-proxy/releases/download/$RELEASE/package-proxy-linux-amd64"

docker build -t="lox24/package-proxy" - << EOF
FROM busybox
ADD $URL /package-proxy
EXPOSE 3142 3143
CMD ["/package-proxy","-dir","/tmp/cache","-tls"]
EOF

docker tag lox24/package-proxy lox24/package-proxy:$TAG
docker push lox24/package-proxy