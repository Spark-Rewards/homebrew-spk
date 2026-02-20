.PHONY: build install clean test release-dry

VERSION ?= dev
COMMIT  := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE    := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -s -w \
	-X github.com/Spark-Rewards/homebrew-spark-cli/cmd.Version=$(VERSION) \
	-X github.com/Spark-Rewards/homebrew-spark-cli/cmd.Commit=$(COMMIT) \
	-X github.com/Spark-Rewards/homebrew-spark-cli/cmd.Date=$(DATE)

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
