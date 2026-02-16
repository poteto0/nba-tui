help:
    @echo "just <recipe>"
    @just --list

default: help

alias ut := unit-test
[group("ci")]
unit-test path=("./...") *args=(""):
    @go test {{path}} {{args}}

alias ut-cov := unit-test-coverage
[group("ci")]
unit-test-coverage path="./..." *args="":
    @go test {{path}} {{args}} -cover -coverprofile=coverage.out

[group("ci")]
lint:
    @golangci-lint run -c .golangci.yaml

[group("ci")]
fmt path=("./..."):
    @go fmt {{path}}

[group("ci")]
fmt-check path=("."):
    @gofmt -l -w {{path}}

# full ci check w/ fmt
[group("ci")]
ci: fmt lint unit-test

[group("build")]
build:
    go build -o nba-tui ./cmd/nba-tui/main.go
