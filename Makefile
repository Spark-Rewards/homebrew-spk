.PHONY: build install clean test release-dry

VERSION ?= dev
COMMIT  := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE    := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -s -w \
	-X github.com/Spark-Rewards/homebrew-spk/cmd.Version=$(VERSION) \
	-X github.com/Spark-Rewards/homebrew-spk/cmd.Commit=$(COMMIT) \
	-X github.com/Spark-Rewards/homebrew-spk/cmd.Date=$(DATE)

build:
	go build -ldflags '$(LDFLAGS)' -o bin/spk main.go

install: build
	cp bin/spk /usr/local/bin/spk
	@echo "spk installed to /usr/local/bin/spk"

clean:
	rm -rf bin/ dist/

test:
	go test ./...

release-dry:
	goreleaser release --snapshot --clean
