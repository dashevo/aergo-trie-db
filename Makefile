# System setup
SHELL = bash

COVERAGE_FILE ?= coverage.txt

.PHONY: default test cover goimports lint vet help clean protoc build check

default:  protoc goimports lint vet build ## Run default target : all lints + test

check_format:  ## Check formatting
	$(eval GOIMPORTS=$(shell goimports -d .))
	@if [ "x$(GOIMPORTS)" != "x" ]; then echo "Go files must be formatted with goimports"; goimports -d -e .; exit 1; fi

check: check_format lint ## Run Go format and lint checks

test:  ## Run a basic test suite
	go test

cover:  ## Run tests and generate test coverage file, output coverage results and HTML coverage file.
	go test -coverprofile $(COVERAGE_FILE)
	go tool cover -func=$(COVERAGE_FILE)
	go tool cover -html=$(COVERAGE_FILE)
	rm -f $(COVERAGE_FILE)

goimports:  ## Run goimports to format code
	goimports -w .

lint:  ## Lint all go code in project
	golint -set_exit_status ./...

vet:  ## Go vet all project code
	go vet ./...

build:  ## Go build
	go build -o bin/server server/*.go
	go build -o bin/client client/*.go

help:  ## Show This Help
	@for line in $$(cat Makefile | grep "##" | grep -v "grep" | sed  "s/:.*##/:/g" | sed "s/\ /!/g"); do verb=$$(echo $$line | cut -d ":" -f 1); desc=$$(echo $$line | cut -d ":" -f 2 | sed "s/!/\ /g"); printf "%-30s--%s\n" "$$verb" "$$desc"; done

clean:  ## Clean up transient (generated) files
	rm -f bin/*
	rm -f $(COVERAGE_FILE)

protoc:  ## Compile protobuf files for Go
	@for D in $$(find . -name \*.proto -print0 | xargs -0 -n 1 dirname | sort -u); do (cd $$D && protoc -I. --go_out=plugins=grpc:. *.proto); done
