.PHONY: *

test:
	go test -v ./...

test-watch:
	watch -n1 go test -v ./...

test-cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	go tool cover -html=coverage.out

build:
	goreleaser release --clean --snapshot --skip=publish

release:
	goreleaser release --clean
