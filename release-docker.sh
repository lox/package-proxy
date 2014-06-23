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
FROM progrium/busybox
ADD $URL /package-proxy
ADD https://raw.githubusercontent.com/bagder/ca-bundle/master/ca-bundle.crt /etc/ssl/ca-bundle.pem
RUN chmod +x /package-proxy
EXPOSE 3142 3143
CMD ["/package-proxy","-dir","/tmp/cache","-tls"]
EOF

docker tag lox24/package-proxy lox24/package-proxy:$TAG
docker push lox24/package-proxy:$TAG

docker tag lox24/package-proxy lox24/package-proxy:latest
docker push lox24/package-proxy:latest