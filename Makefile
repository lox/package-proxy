#!/usr/bin/make -f

GIT_VERSION := $(shell git describe --tags --abbrev=4 --always --dirty)

clean:
	-rm -rf release
	-rm -rf packageproxy

build: deps
	go build .

release: deps
	@mkdir -p release/${GIT_VERSION}/linux-amd64 release/${GIT_VERSION}/darwin-amd64
	GOARCH=amd64 GOOS=linux  go build -o release/${GIT_VERSION}/linux-amd64/package-proxy
	GOARCH=amd64 GOOS=darwin go build -o release/${GIT_VERSION}/darwin-amd64/package-proxy
	zip -D release/${GIT_VERSION}_linux-amd64.zip release/${GIT_VERSION}/linux-amd64/package-proxy
	zip -D release/${GIT_VERSION}_darwin-amd64.zip release/${GIT_VERSION}/darwin-amd64/package-proxy

deps:
	go get -u

docker: clean
	GOARCH=amd64 GOOS=linux go build -o package-proxy-linux-amd64
	docker build --tag="package-proxy" .
	-rm package-proxy-linux-amd64

docker-run:
	docker run \
		--tty --interactive --rm --publish 3142:3142 \
		--volume /tmp/cache:/tmp/cache \
		--volume /projects/go:/gopath \
		package-proxy

docker-release: docker
	docker tag package-proxy lox24/package-proxy:latest
	docker push lox24/package-proxy:latest