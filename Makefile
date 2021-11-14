VERSION ?= latest

.DEFAULT_TARGET = help

## help: Display list of commands
.PHONY: help
help: Makefile
	@sed -n 's|^##||p' $< | column -t -s ':' | sed -e 's|^| |'

## test: Run tests
test: fmt vet
	go test ./... -coverprofile cover.out

## build: Build a container image
.PHONY: build
build: fmt vet
	go build -o bin/secret-santa main.go

## run: Run the CLI
.PHONY: run
run: fmt vet
	go run main.go

## fmt: Format the source code
.PHONY: fmt
fmt:
	go fmt ./...

## vet: Vet the source code
.PHONY: vet
vet:
	go vet ./...
