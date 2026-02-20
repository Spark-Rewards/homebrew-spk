.PHONY: build install clean test release-dry

# Release version: single source of truth is the VERSION file in repo root
VERSION ?= $(shell cat VERSION 2>/dev/null | tr -d '[:space:]' || echo "dev")
COMMIT  := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE    := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -s -w \
	-X github.com/Spark-Rewards/homebrew-spk/cmd.Version=$(VERSION) \
	-X github.com/Spark-Rewards/homebrew-spk/cmd.Commit=$(COMMIT) \
	-X github.com/Spark-Rewards/homebrew-spk/cmd.Date=$(DATE)

build:
	go build -ldflags '$(LDFLAGS)' -o bin/spark-cli main.go

install: build
	cp bin/spark-cli /usr/local/bin/spark-cli
	@echo "spark-cli installed to /usr/local/bin/spark-cli"

clean:
	rm -rf bin/ dist/

test:
	go test ./...

release-dry:
	goreleaser release --snapshot --clean
