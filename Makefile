#!/usr/bin/make -f

NAME := package-proxy
VERSION := $(shell git describe --tags --abbrev=4 --always)
DOCKER_VERSION := $(subst v,,$(VERSION))
BUILD_ARCHS := amd64
BUILD_OOS := linux darwin
BUILDS = $(foreach oos, $(BUILD_OOS), \
	$(foreach arch, $(BUILD_ARCHS), $(oos)/$(arch)))

.PHONY: clean build release docker docker-run docker-release

clean:
	-rm -rf bin/ release/ package-proxy

build: $(foreach build, $(BUILDS), bin/$(build)/$(NAME))

bin/%/$(NAME): deps
	@mkdir -p bin/$*
	GOOS=$(firstword $(subst /, ,$*)) GOARCH=$(lastword $(subst /, ,$*)) go build -o bin/$*/$(NAME)

release: build $(foreach build, $(BUILDS), release/$(VERSION)/$(build))

release/$(VERSION)/%: release/$(VERSION)
	github-release upload --user lox --repo $(NAME) --tag $(VERSION) \
		--name "$(NAME)-$(subst /,-,$*)" --file bin/$(*)/$(NAME)
	cp bin/$(*)/$(NAME) release/$(VERSION)/$(subst /,-,$*)

release/$(VERSION):
	github-release release --user lox --repo $(NAME) --tag $(VERSION)
	@mkdir -p release/$(VERSION)

deps:
	go get

docker: clean build
	docker build --tag="lox24/package-proxy" .

docker-run:
	docker run \
		--name package-proxy \
		--detach \
		--publish 3142:3142 \
		--volume /tmp/vagrant-cache/generic:/tmp/cache \
		lox24/package-proxy

docker-release:
	docker tag lox24/package-proxy:latest lox24/package-proxy:$(DOCKER_VERSION)
	docker push lox24/package-proxy:$(DOCKER_VERSION)