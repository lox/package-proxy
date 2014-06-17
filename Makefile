#!/usr/bin/make -f

NAME := package-proxy
VERSION := $(shell git describe --tags --abbrev=4 --always)
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
	go build -o bin/$*/$(NAME)

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
	docker build --tag="package-proxy" .

docker-run: docker
	docker run \
		--tty --interactive --rm --publish 3142:3142 \
		--volume /tmp/vagrant-cache/generic:/tmp/cache \
		package-proxy

docker-release: docker
	docker tag package-proxy lox24/package-proxy:$(subst v,,$(VERSION))
	docker push lox24/package-proxy:$(subst v,,$(VERSION))
	docker tag package-proxy lox24/package-proxy:latest
	docker push lox24/package-proxy:latest