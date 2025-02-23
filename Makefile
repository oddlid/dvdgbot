BINARY := dvdgbot.bin
VERSION := 2025-02-23
UNAME := $(shell uname -s)
SOURCES := $(wildcard *.go)
DEPS := $(wildcard leet/*.go larsmonsen/*.go xkcdbot/*.go userwatch/*.go timestamp/*.go quoteshuffle/*.go morse/*.go)
COMMIT_ID := $(shell git describe --tags --always)
BUILD_TIME := $(shell go run tool/rfc3339date.go)
LDFLAGS = -ldflags "-X main.Version=${VERSION} -X main.BuildDate=${BUILD_TIME} -X main.CommitID=${COMMIT_ID} -X main.BinaryName=${BINARY} -s -w ${DFLAG}"

ifeq ($(UNAME), Linux)
	DFLAG := -d
endif

.DEFAULT_GOAL: $(BINARY)

$(BINARY): $(SOURCES) $(DEPS)
	env CGO_ENABLED=0 go build ${LDFLAGS} -o $@ .

.PHONY: run
run:
	env DEBUG=1 go run .

.PHONY: install
install:
	env CGO_ENABLED=0 go install ${LDFLAGS} ./...

.PHONY: clean
clean:
	if [ -f ${BINARY} ]; then rm -f ${BINARY}; fi
