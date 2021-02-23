BINARY := bajsbot.bin
VERSION := 2021-02-23
SOURCES := $(wildcard *.go)
DEPS := $(wildcard leet/*.go larsmonsen/*.go xkcdbot/*.go goodmorning/*.go quote/*.go userwatch/*.go timestamp/*.go quoteshuffle/*.go)
COMMIT_ID := $(shell git describe --tags --always)
BUILD_TIME := $(shell go run tool/rfc3339date.go)
LDFLAGS = -ldflags "-X main.VERSION=${VERSION} -X main.BUILD_DATE=${BUILD_TIME} -X main.COMMIT_ID=${COMMIT_ID} -s -w ${DFLAG}"

ifeq ($(UNAME), Linux)
	DFLAG := -d
endif

.DEFAULT_GOAL: $(BINARY)

$(BINARY): $(SOURCES) $(DEPS)
	env GO111MODULE=off CGO_ENABLED=0 go build ${LDFLAGS} -o $@ .

.PHONY: install
install:
	env GO111MODULE=off CGO_ENABLED=0 go install ${LDFLAGS} ./...

.PHONY: clean
clean:
	if [ -f ${BINARY} ]; then rm -f ${BINARY}; fi
