FROM ubuntu:14.04
ENV GOPATH /go
RUN apt-get -y update --no-install-recommends
RUN apt-get -y install --no-install-recommends golang-go bzr git ca-certificates
RUN go get github.com/aktau/github-release
RUN go get github.com/lox/package-proxy
ADD release-docker.sh /release-docker.sh
ENV GOBIN /go/bin
ENV PATH $GOBIN:$PATH
WORKDIR /go/src/github.com/lox
ENTRYPOINT ["/go/bin/package-proxy","-dir","/tmp/cache"]