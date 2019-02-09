BINARY := bajsbot.bin
VERSION := 2019-02-09
SOURCES := $(wildcard *.go)
DEPS := larsmonsen/larsmonsen.go leet/leet.go xkcdbot/xkcdbot.go goodmorning/goodmorning.go quote/quote.go userwatch/userwatch.go
COMMIT_ID := $(shell git describe --tags --always)
BUILD_TIME := $(shell date +%FT%T%:z)
LDFLAGS = -ldflags "-X main.VERSION=${VERSION} -X main.BUILD_DATE=${BUILD_TIME} -X main.COMMIT_ID=${COMMIT_ID} -d -s -w"

.DEFAULT_GOAL: $(BINARY)

$(BINARY): $(SOURCES) $(DEPS)
	env CGO_ENABLED=0 go build ${LDFLAGS} -o $@ ${SOURCES}

.PHONY: install
install:
	env CGO_ENABLED=0 go install ${LDFLAGS} ./...

.PHONY: clean
clean:
	if [ -f ${BINARY} ]; then rm -f ${BINARY}; fi
