# Display this help message
.PHONY: help
help:
	@awk '/^.PHONY:/ && (a ~ /#/) {gsub(/.PHONY: /, "", $$0); gsub(/# /, "", a); printf "\033[0;32m%-10s\033[0m%s\n", $$0, a}{a=$$0}' $(MAKEFILE_LIST)

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
