# Display this help message
.PHONY: help
help:
	@awk '/^.PHONY:/ && (a ~ /#/) {gsub(/.PHONY: /, "", $$0); gsub(/# /, "", a); printf "\033[0;32m%-15s\033[0m%s\n", $$0, a}{a=$$0}' $(MAKEFILE_LIST)

# Check code for errors
.PHONY: check
check:
	go vet ./...

# Run unit tests
.PHONY: test
test: check
	go test -v -race ./...

# Compile into executable binary
.PHONY: build
build: test
	CGO_ENABLED=0 go build -o stars -ldflags "-X main.Version=$(shell cat VERSION)" ./cmd/stars

# Do a release. VERSION needs to be bumped manually
.PHONY: release
release:
	git tag v$(shell cat VERSION)
	git push origin master
	goreleaser --rm-dist

# Do a major release
.PHONY: release-major
release-major:
	@echo $(shell awk -F. '{printf "%s.%s.%s\n", $$1+1, $$2, $$3}' VERSION) > VERSION
	@git add VERSION
	@git commit -S -m "Release $(shell cat VERSION)"
	$(MAKE) release

# Do a minor release
.PHONY: release-minor
release-minor:
	@echo $(shell awk -F. '{printf "%s.%s.%s\n", $$1, $$2+1, $$3}' VERSION) > VERSION
	@git add VERSION
	@git commit -S -m "Release $(shell cat VERSION)"
	$(MAKE) release

# Do a patch release
.PHONY: release-patch
release-patch:
	@echo $(shell awk -F. '{printf "%s.%s.%s\n", $$1, $$2, $$3+1}' VERSION) > VERSION
	@git add VERSION
	@git commit -S -m "Release $(shell cat VERSION)"
	$(MAKE) release
