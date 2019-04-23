.PHONY: check
check:
	go vet ./...

.PHONY: test
test: check
	go test -v -race ./...

.PHONY: build
build: test
	CGO_ENABLED=0 go build -o stars -ldflags "-X main.Version=$(shell cat VERSION)" ./cmd

.PHONY: release
release:
	git tag v$(shell cat VERSION)
	git push origin master
	goreleaser --rm-dist